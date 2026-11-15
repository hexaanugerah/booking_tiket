#!/usr/bin/env bash
# Menyalakan seluruh stack agar bisa diakses dari luar rumah:
#   1. backend Go      (port 8080)
#   2. frontend Next   (port 3000)
#   3. satu tunnel Cloudflare ke frontend -> URL HTTPS publik
#
# Semua request API/berkas lewat tunnel itu juga: server Next meneruskan
# /api/v1/* dan /uploads/* ke backend lewat rewrites (next.config.ts). Jadi
# tidak perlu tunnel terpisah untuk backend, tidak perlu build/deploy Worker,
# dan tidak ada CORS karena browser hanya bicara ke origin halamannya sendiri.
#
# Pakai:  bash scripts/start-public.sh
#
# Prasyarat: MySQL + Redis sudah jalan (backend butuh keduanya) dan port 8080
# serta 3000 bebas. Tunnel hanya butuh koneksi keluar, jadi tidak perlu buka
# port di router/firewall.
#
# Catatan: URL quick tunnel berubah setiap kali skrip dijalankan. Kalau butuh
# URL tetap, pakai named tunnel Cloudflare (butuh akun + domain).

set -uo pipefail

# Windows Dev Disco menaruh Go di "c:\Program Files\Go\bin", yang tidak selalu
# masuk ke PATH bash di MSYS2/Git Bash. Jika go ternyata tidak ada, coba tambah
# lokasi tersebut sebelum build / tunnel.
if ! command -v go >/dev/null 2>&1; then
  for p in \
    "/c/Program Files/Go/bin" \
    "/c/Program Files (x86)/Go/bin" \
    "$HOME/go/bin" \
    "/usr/local/go/bin" \
    ; do
    if [ -d "$p" ] && [ -x "$p/go" ]; then
      export PATH="$p:$PATH"
      break
    fi
  done
fi

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
BACKEND="$ROOT/backend"
FRONTEND="$ROOT/frontend"
LOGDIR="${TMPDIR:-/tmp}/kawahputih"
mkdir -p "$LOGDIR"

BACKEND_PORT="${APP_PORT:-8080}"
FRONTEND_PORT="${FRONTEND_PORT:-3000}"

case "$(uname -s)" in
  MINGW*|MSYS*|CYGWIN*) SERVER_BIN="bin/server.exe" ;;
  *) SERVER_BIN="bin/server" ;;
esac

say() { printf '\n== %s\n' "$*"; }

wait_http() { # url, percobaan, jeda detik
  local url="$1" tries="$2" delay="$3"
  for _ in $(seq 1 "$tries"); do
    if curl -s -o /dev/null --max-time 5 "$url"; then return 0; fi
    sleep "$delay"
  done
  return 1
}

# --- 1. Backend ------------------------------------------------------------
say "backend (:$BACKEND_PORT)"
if curl -s -o /dev/null --max-time 3 "http://127.0.0.1:$BACKEND_PORT/healthz"; then
  echo "sudah jalan, dilewati"
else
  if [ ! -x "$BACKEND/$SERVER_BIN" ]; then
    echo "binary belum ada, build dulu (go build)"
    ( cd "$BACKEND" && go build -o "$SERVER_BIN" ./cmd/server ) \
      || { echo "GAGAL: build backend" >&2; exit 1; }
  fi
  ( cd "$BACKEND" && nohup "./$SERVER_BIN" > "$LOGDIR/backend.log" 2>&1 & )
  if ! wait_http "http://127.0.0.1:$BACKEND_PORT/healthz" 30 2; then
    echo "GAGAL: backend tidak siap. Cek $LOGDIR/backend.log (MySQL/Redis sudah jalan?)" >&2
    exit 1
  fi
  curl -s -o /dev/null -w 'healthz=%{http_code}\n' --max-time 5 "http://127.0.0.1:$BACKEND_PORT/healthz"
fi

# --- 2. Frontend -----------------------------------------------------------
say "frontend (:$FRONTEND_PORT)"
if curl -s -o /dev/null --max-time 5 "http://127.0.0.1:$FRONTEND_PORT/"; then
  echo "sudah jalan, dilewati"
else
  ( cd "$FRONTEND" && nohup npm run dev -- -p "$FRONTEND_PORT" > "$LOGDIR/frontend.log" 2>&1 & )
  if ! wait_http "http://127.0.0.1:$FRONTEND_PORT/" 60 4; then
    echo "GAGAL: frontend tidak siap. Cek $LOGDIR/frontend.log" >&2
    exit 1
  fi
  curl -s -o /dev/null -w 'home=%{http_code}\n' --max-time 15 "http://127.0.0.1:$FRONTEND_PORT/"
fi

# --- 3. Tunnel ke frontend -------------------------------------------------
say "tunnel publik (:$FRONTEND_PORT -> HTTPS)"
: > "$LOGDIR/tunnel.log"
npx -y cloudflared@latest tunnel --url "http://localhost:$FRONTEND_PORT" --no-autoupdate \
  > "$LOGDIR/tunnel.log" 2>&1 &
TUNNEL_PID=$!

PUBLIC_URL=""
for _ in $(seq 1 75); do
  sleep 2
  # npx/npm sementara node di MSYS2 sering minta stdio saat instalasi paket.
  # Cek apakah proses masih hidup; kalau sudah exit, berhenti polling.
  if ! kill -0 "$TUNNEL_PID" 2>/dev/null; then
    # baca tail log terakhir sebelum keluar, agar pesan error terlihat
    tail -n 20 "$LOGDIR/tunnel.log"
    break
  fi
  PUBLIC_URL=$(grep -hoE 'https://[a-z0-9-]+(\.trycloudflare\.com|\.ts\.net)' "$LOGDIR/tunnel.log" | tail -1)
  [ -n "$PUBLIC_URL" ] && break
done

if [ -z "$PUBLIC_URL" ]; then
  echo "GAGAL: URL tunnel tidak muncul. Lihat $LOGDIR/tunnel.log" >&2
  exit 1
fi
echo "$PUBLIC_URL" > "$LOGDIR/public-url.txt"
echo "URL publik: $PUBLIC_URL"

# biarkan tunnel jalan sendiri; jangan tunggu sampai selesai
disown "$TUNNEL_PID" 2>/dev/null
wait "$TUNNEL_PID" 2>/dev/null

# --- 4. Verifikasi lewat URL publik ---------------------------------------
say "verifikasi dari internet"
home=$(curl -s -o /dev/null -w '%{http_code}' --max-time 45 "$PUBLIC_URL/")
api=$(curl -s -o /dev/null -w '%{http_code}' --max-time 45 "$PUBLIC_URL/api/v1/faq")
echo "halaman=$home  api=$api"
if [ "$home" != "200" ] || [ "$api" != "200" ]; then
  echo "PERINGATAN: hasil belum 200. Tunggu beberapa detik lalu coba buka $PUBLIC_URL" >&2
fi

cat <<EOF

=========================== RINGKASAN ===========================
URL publik     : $PUBLIC_URL   <- buka dari HP/jaringan mana pun
Frontend lokal : http://localhost:$FRONTEND_PORT
Backend lokal  : http://localhost:$BACKEND_PORT
LAN            : http://<IP-komputer>:$FRONTEND_PORT
Log            : $LOGDIR

Proses tetap jalan setelah terminal ditutup.
Matikan semua  : taskkill //F //IM cloudflared.exe; taskkill //F //IM ${SERVER_BIN##*/}
                 (Linux/macOS: pkill cloudflared; pkill -f "${SERVER_BIN##*/}")
=================================================================
EOF
