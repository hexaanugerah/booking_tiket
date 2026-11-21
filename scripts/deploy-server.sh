#!/usr/bin/env bash
# =============================================================================
# Deploy otomatis Kawah Putih ke server Ubuntu/Debian.
#
# Cukup jalankan dari root projek:
#   sudo bash scripts/deploy-server.sh
#
# Di server baru, cara tercepat (clone + deploy sekaligus):
#   curl -fsSL https://raw.githubusercontent.com/hexaanugerah/booking_tiket/main/scripts/bootstrap.sh | sudo bash
#
# Tanpa konfigurasi tambahan, aplikasi langsung bisa dibuka lewat IP server
# (http://IP_SERVER) tanpa domain, DNS, atau panel web apa pun.
#
# Untuk memakai domain + HTTPS:
#   1. cp deploy.env.example deploy.env
#   2. isi DOMAIN, ENABLE_HTTPS=true, dan LETSENCRYPT_EMAIL
#   3. sudo bash scripts/deploy-server.sh
#
# Skrip ini idempotent: aman dijalankan ulang untuk update/rebuild.
# =============================================================================
set -Eeuo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" >/dev/null 2>&1 && pwd)"
ROOT_DIR="$(cd -- "$SCRIPT_DIR/.." >/dev/null 2>&1 && pwd)"

# --- Muat konfigurasi opsional ----------------------------------------------
if [[ -f "$ROOT_DIR/deploy.env" ]]; then
  set -a
  # shellcheck disable=SC1091
  source "$ROOT_DIR/deploy.env"
  set +a
fi

DOMAIN="${DOMAIN:-}"
ENABLE_HTTPS="${ENABLE_HTTPS:-false}"
LETSENCRYPT_EMAIL="${LETSENCRYPT_EMAIL:-}"
PUBLIC_HOST_OVERRIDE="${PUBLIC_HOST:-}"
APP_DIR="${APP_DIR:-/opt/booking_tiket}"
REPO_URL="${REPO_URL:-https://github.com/hexaanugerah/booking_tiket.git}"
APP_ENV="${APP_ENV:-local}"
STOP_OLD_STACK="${STOP_OLD_STACK:-true}"
ENABLE_UFW="${ENABLE_UFW:-true}"
APP_PORT="${APP_PORT:-8080}"
# auto  : pakai Nginx host bila port 80 bebas; jika tidak, buka APP_PORT langsung.
# true  : paksa Nginx host (port 80/443).
# false : tanpa Nginx host, akses langsung lewat http://host:APP_PORT.
EDGE_PROXY="${EDGE_PROXY:-auto}"

MYSQL_ROOT_PASSWORD="${MYSQL_ROOT_PASSWORD:-}"
DB_NAME="${DB_NAME:-kawah_putih_db}"
DB_USER="${DB_USER:-kawahputih}"
DB_PASSWORD="${DB_PASSWORD:-}"
JWT_ACCESS_SECRET="${JWT_ACCESS_SECRET:-}"
JWT_REFRESH_SECRET="${JWT_REFRESH_SECRET:-}"
JWT_ACCESS_TTL_MINUTES="${JWT_ACCESS_TTL_MINUTES:-15}"
JWT_REFRESH_TTL_DAYS="${JWT_REFRESH_TTL_DAYS:-7}"
JWT_ISSUER="${JWT_ISSUER:-KawahPutih}"
MIDTRANS_SERVER_KEY="${MIDTRANS_SERVER_KEY:-}"
MIDTRANS_IS_PRODUCTION="${MIDTRANS_IS_PRODUCTION:-false}"
SMTP_HOST="${SMTP_HOST:-}"
SMTP_PORT="${SMTP_PORT:-587}"
SMTP_USER="${SMTP_USER:-}"
SMTP_PASSWORD="${SMTP_PASSWORD:-}"
SMTP_FROM="${SMTP_FROM:-no-reply@kawahputih.id}"
GOOGLE_CLIENT_ID="${GOOGLE_CLIENT_ID:-}"
GOOGLE_CLIENT_SECRET="${GOOGLE_CLIENT_SECRET:-}"
GOOGLE_REDIRECT_URL="${GOOGLE_REDIRECT_URL:-}"
SEED_SUPERADMIN_PASSWORD="${SEED_SUPERADMIN_PASSWORD:-}"
SEED_ADMIN_PASSWORD="${SEED_ADMIN_PASSWORD:-}"
SEED_MANAGER_PASSWORD="${SEED_MANAGER_PASSWORD:-}"
SEED_PETUGAS_PASSWORD="${SEED_PETUGAS_PASSWORD:-}"
SEED_RESET_STAFF_PASSWORD="${SEED_RESET_STAFF_PASSWORD:-false}"

# --- Output helpers ----------------------------------------------------------
if [[ -t 1 ]]; then
  C_BLUE=$'\033[1;34m'; C_YELLOW=$'\033[1;33m'; C_RED=$'\033[1;31m'; C_GREEN=$'\033[1;32m'; C_RESET=$'\033[0m'
else
  C_BLUE=""; C_YELLOW=""; C_RED=""; C_GREEN=""; C_RESET=""
fi
log()  { printf '\n%s==> %s%s\n' "$C_BLUE" "$*" "$C_RESET"; }
warn() { printf '%s[!] %s%s\n' "$C_YELLOW" "$*" "$C_RESET"; }
ok()   { printf '%s[ok] %s%s\n' "$C_GREEN" "$*" "$C_RESET"; }
die()  { printf '%s[x] %s%s\n' "$C_RED" "$*" "$C_RESET" >&2; exit 1; }
rand_hex() { openssl rand -hex "${1:-16}"; }

[[ "${EUID}" -eq 0 ]] || die "Jalankan sebagai root: sudo bash scripts/deploy-server.sh"
command -v apt-get >/dev/null 2>&1 || die "Skrip ini untuk Ubuntu/Debian (apt-get)."

# =============================================================================
# 1. Dependensi sistem
# =============================================================================
log "1/8 Memasang dependensi sistem"
export DEBIAN_FRONTEND=noninteractive
apt-get update -y
apt-get install -y ca-certificates curl git openssl iproute2

if ! command -v docker >/dev/null 2>&1; then
  log "Memasang Docker"
  curl -fsSL https://get.docker.com | sh
fi
systemctl enable --now docker >/dev/null 2>&1 || true
docker compose version >/dev/null 2>&1 || die "Docker Compose plugin tidak tersedia."
ok "Docker $(docker --version | awk '{print $3}' | tr -d ',')"

# =============================================================================
# 2. Kode projek
# =============================================================================
log "2/8 Menyiapkan kode projek"
if [[ -f "$ROOT_DIR/docker-compose.yml" && -f "$ROOT_DIR/Dockerfile" ]]; then
  PROJECT_DIR="$ROOT_DIR"
else
  if [[ -d "$APP_DIR/.git" ]]; then
    git -c safe.directory="$APP_DIR" -C "$APP_DIR" pull --ff-only \
      || warn "git pull gagal; melanjutkan dengan kode yang sudah ada."
  else
    mkdir -p "$(dirname "$APP_DIR")"
    git clone "$REPO_URL" "$APP_DIR"
  fi
  if [[ -f "$APP_DIR/kawahputih/docker-compose.yml" ]]; then
    PROJECT_DIR="$APP_DIR/kawahputih"
  elif [[ -f "$APP_DIR/docker-compose.yml" ]]; then
    PROJECT_DIR="$APP_DIR"
  else
    die "docker-compose.yml tidak ditemukan. Pastikan perubahan Dockerfile sudah di-push ke GitHub."
  fi
fi
cd "$PROJECT_DIR"
[[ -f Dockerfile ]] || die "Dockerfile tidak ditemukan di $PROJECT_DIR"
ok "Projek: $PROJECT_DIR"

# =============================================================================
# 3. Hentikan stack lama yang memakai port yang sama
# =============================================================================
log "3/8 Memeriksa stack lama"
OLD_COMPOSE="$PROJECT_DIR/docker/docker-compose.prod.yml"
OLD_BACKUP_FILE=""
if [[ "$STOP_OLD_STACK" == "true" && -f "$OLD_COMPOSE" ]]; then
  if docker compose -f "$OLD_COMPOSE" ps -q 2>/dev/null | grep -q .; then
    mkdir -p "$PROJECT_DIR/backups"
    OLD_BACKUP_FILE="$PROJECT_DIR/backups/db-sebelum-deploy-$(date +%Y%m%d-%H%M%S).sql"
    warn "Membuat backup database stack lama sebelum dihentikan..."
    if docker compose -f "$OLD_COMPOSE" exec -T mysql sh -c \
      'exec mysqldump -uroot -p"$MYSQL_ROOT_PASSWORD" --single-transaction --routines --triggers "$MYSQL_DATABASE"' \
      > "$OLD_BACKUP_FILE"; then
      ok "Backup database: $OLD_BACKUP_FILE"
    else
      rm -f "$OLD_BACKUP_FILE"
      OLD_BACKUP_FILE=""
      die "Backup database lama gagal. Stack lama TIDAK dihentikan agar data tetap aman."
    fi
    warn "Menghentikan stack lama (volume/data tetap aman)..."
    docker compose -f "$OLD_COMPOSE" down
  fi
fi

port_in_use() {
  ss -ltn 2>/dev/null | awk 'NR > 1 { print $4 }' | grep -Eq "[:.]$1$"
}
if port_in_use "$APP_PORT"; then
  PORT_OWNER="$(docker ps --format '{{.Names}} {{.Ports}}' 2>/dev/null | awk -v p=":$APP_PORT->" 'index($0, p) { print $1; exit }')"
  if [[ -n "$PORT_OWNER" ]]; then
    OWNER_PROJECT="$(docker inspect -f '{{ index .Config.Labels "com.docker.compose.project" }}' "$PORT_OWNER" 2>/dev/null || true)"
    if [[ "$OWNER_PROJECT" == "kawahputih" ]]; then
      ok "Port $APP_PORT sedang dipakai stack Kawah Putih; container akan di-recreate."
    else
      die "Port $APP_PORT dipakai container '$PORT_OWNER' (project: ${OWNER_PROJECT:-bukan compose}). Hentikan dulu."
    fi
  else
    die "Port $APP_PORT sudah dipakai proses non-Docker. Hentikan proses tersebut atau ubah APP_PORT di deploy.env."
  fi
fi

# Tentukan mode akses. Bila port 80 sudah dipakai proses lain (Coolify,
# Traefik, Apache, dsb.), jangan pasang Nginx; buka gateway langsung di
# APP_PORT supaya tidak bentrok.
if [[ "$EDGE_PROXY" == "auto" ]]; then
  if port_in_use 80; then
    if [[ -e /etc/nginx/sites-enabled/kawahputih ]] && systemctl is-active --quiet nginx; then
      EDGE_PROXY="true"
    else
      EDGE_PROXY="false"
    fi
  else
    EDGE_PROXY="true"
  fi
fi
case "$EDGE_PROXY" in
  true)
    if port_in_use 80 \
      && ! { [[ -e /etc/nginx/sites-enabled/kawahputih ]] && systemctl is-active --quiet nginx; }; then
      die "EDGE_PROXY=true tetapi port 80 dipakai proses lain (Coolify/Traefik/Apache). Gunakan EDGE_PROXY=auto atau false."
    fi
    STACK_BIND_HOST="127.0.0.1"
    ;;
  false)
    ENABLE_HTTPS="false"
    STACK_BIND_HOST="0.0.0.0"
    warn "Port 80 dipakai proses lain (kemungkinan Coolify/Traefik). Nginx host dilewati."
    ;;
  *)
    die "EDGE_PROXY harus auto, true, atau false."
    ;;
esac

# =============================================================================
# 4. Nginx (dan HTTPS bila diminta)
# =============================================================================
log "4/8 Mengatur reverse proxy"
SCHEME="http"
if [[ "$EDGE_PROXY" == "true" ]]; then
  apt-get install -y nginx
  if [[ "$ENABLE_HTTPS" == "true" ]]; then
    apt-get install -y certbot python3-certbot-nginx
  fi

if [[ -n "$DOMAIN" ]]; then
  SERVER_NAME="$DOMAIN"
  LISTEN_80="listen 80;"
  LISTEN_80_V6="listen [::]:80;"
else
  SERVER_NAME="_"
  LISTEN_80="listen 80 default_server;"
  LISTEN_80_V6="listen [::]:80 default_server;"
fi

NGINX_SITE="/etc/nginx/sites-available/kawahputih"
if [[ -f "$NGINX_SITE" ]]; then
  cp -a "$NGINX_SITE" "${NGINX_SITE}.bak.$(date +%Y%m%d%H%M%S)"
fi

write_nginx_http() {
  cat > "$NGINX_SITE" <<NGINX
server {
    $LISTEN_80
    $LISTEN_80_V6
    server_name $SERVER_NAME;

    client_max_body_size 25m;

    # Semua request diteruskan ke gateway container. Gateway yang membagi
    # /api/v1 dan /uploads ke backend, sisanya ke frontend.
    location / {
        proxy_pass http://127.0.0.1:$APP_PORT;
        proxy_http_version 1.1;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_read_timeout 120s;
    }
}
NGINX
}

write_nginx_http
ln -sf "$NGINX_SITE" /etc/nginx/sites-enabled/kawahputih
if [[ -e /etc/nginx/sites-enabled/default ]]; then
  rm -f /etc/nginx/sites-enabled/default
fi

nginx -t
systemctl enable --now nginx >/dev/null 2>&1 || true
systemctl reload nginx
ok "Nginx aktif di port 80"
else
  warn "Mode langsung: aplikasi dibuka di http://<host>:$APP_PORT (tanpa Nginx/HTTPS)."
fi

if [[ "$EDGE_PROXY" == "true" && -n "$DOMAIN" && "$ENABLE_HTTPS" == "true" ]]; then
  log "Menerbitkan sertifikat HTTPS untuk $DOMAIN"
  RESOLVED_IP="$(getent ahostsv4 "$DOMAIN" 2>/dev/null | awk 'NR == 1 { print $1 }' || true)"
  SERVER_IP="$(curl -4 -fsS --max-time 5 https://api.ipify.org 2>/dev/null || true)"
  if [[ -n "$RESOLVED_IP" && -n "$SERVER_IP" && "$RESOLVED_IP" != "$SERVER_IP" ]]; then
    warn "DNS $DOMAIN mengarah ke $RESOLVED_IP, sedangkan IP publik server $SERVER_IP."
  fi

  CERTBOT_ARGS=(--nginx --non-interactive --agree-tos --redirect -d "$DOMAIN")
  if [[ -n "$LETSENCRYPT_EMAIL" ]]; then
    CERTBOT_ARGS+=(--email "$LETSENCRYPT_EMAIL")
  else
    CERTBOT_ARGS+=(--register-unsafely-without-email)
  fi

  if certbot "${CERTBOT_ARGS[@]}"; then
    SCHEME="https"
    ok "HTTPS aktif"
  else
    warn "Gagal menerbitkan sertifikat (biasanya DNS belum mengarah ke server atau port 80 tertutup)."
    warn "Aplikasi tetap dijalankan lewat HTTP."
    write_nginx_http
    nginx -t
    systemctl reload nginx
  fi
elif [[ "$EDGE_PROXY" == "true" && -n "$DOMAIN" ]]; then
  warn "ENABLE_HTTPS=false: aplikasi memakai http://$DOMAIN"
fi

# =============================================================================
# 5. Tentukan alamat publik
# =============================================================================
PUBLIC_HOST="${PUBLIC_HOST_OVERRIDE:-$DOMAIN}"
if [[ -z "$PUBLIC_HOST" ]]; then
  PUBLIC_HOST="$(curl -4 -fsS --max-time 5 https://api.ipify.org 2>/dev/null || true)"
fi
if [[ -z "$PUBLIC_HOST" ]]; then
  PUBLIC_HOST="$(hostname -I 2>/dev/null | awk '{ print $1 }' || true)"
fi
[[ -n "$PUBLIC_HOST" ]] || die "Tidak bisa menentukan alamat publik. Isi PUBLIC_HOST di deploy.env."
if [[ "$EDGE_PROXY" == "true" ]]; then
  PUBLIC_BASE_URL="$SCHEME://$PUBLIC_HOST"
else
  PUBLIC_BASE_URL="http://$PUBLIC_HOST:$APP_PORT"
fi

# =============================================================================
# 6. Buat/perbarui .env dan rahasia
# =============================================================================
log "5/8 Menyiapkan environment"
ENV_CREATED="false"
if [[ ! -f .env ]]; then
  MYSQL_ROOT_PASSWORD="${MYSQL_ROOT_PASSWORD:-$(rand_hex 24)}"
  DB_PASSWORD="${DB_PASSWORD:-$(rand_hex 24)}"
  JWT_ACCESS_SECRET="${JWT_ACCESS_SECRET:-$(rand_hex 48)}"
  JWT_REFRESH_SECRET="${JWT_REFRESH_SECRET:-$(rand_hex 48)}"
  SEED_SUPERADMIN_PASSWORD="${SEED_SUPERADMIN_PASSWORD:-$(rand_hex 8)}"
  SEED_ADMIN_PASSWORD="${SEED_ADMIN_PASSWORD:-$(rand_hex 8)}"
  SEED_MANAGER_PASSWORD="${SEED_MANAGER_PASSWORD:-$(rand_hex 8)}"
  SEED_PETUGAS_PASSWORD="${SEED_PETUGAS_PASSWORD:-$(rand_hex 8)}"

  umask 077
  cat > .env <<ENVFILE
# Dibuat otomatis oleh scripts/deploy-server.sh pada $(date -Is)
APP_ENV=$APP_ENV
FRONTEND_URL=$PUBLIC_BASE_URL
BACKEND_PUBLIC_URL=$PUBLIC_BASE_URL
NEXT_PUBLIC_API_BASE_URL=$PUBLIC_BASE_URL/api/v1
BIND_HOST=$STACK_BIND_HOST
APP_PORT=$APP_PORT

MYSQL_ROOT_PASSWORD=$MYSQL_ROOT_PASSWORD
DB_NAME=$DB_NAME
DB_USER=$DB_USER
DB_PASSWORD=$DB_PASSWORD

JWT_ACCESS_SECRET=$JWT_ACCESS_SECRET
JWT_REFRESH_SECRET=$JWT_REFRESH_SECRET
JWT_ACCESS_TTL_MINUTES=$JWT_ACCESS_TTL_MINUTES
JWT_REFRESH_TTL_DAYS=$JWT_REFRESH_TTL_DAYS
JWT_ISSUER=$JWT_ISSUER

MIDTRANS_SERVER_KEY=$MIDTRANS_SERVER_KEY
MIDTRANS_IS_PRODUCTION=$MIDTRANS_IS_PRODUCTION

SMTP_HOST=$SMTP_HOST
SMTP_PORT=$SMTP_PORT
SMTP_USER=$SMTP_USER
SMTP_PASSWORD=$SMTP_PASSWORD
SMTP_FROM=$SMTP_FROM

GOOGLE_CLIENT_ID=$GOOGLE_CLIENT_ID
GOOGLE_CLIENT_SECRET=$GOOGLE_CLIENT_SECRET
GOOGLE_REDIRECT_URL=${GOOGLE_REDIRECT_URL:-$PUBLIC_BASE_URL/api/v1/auth/google/callback}

SEED_SUPERADMIN_PASSWORD=$SEED_SUPERADMIN_PASSWORD
SEED_ADMIN_PASSWORD=$SEED_ADMIN_PASSWORD
SEED_MANAGER_PASSWORD=$SEED_MANAGER_PASSWORD
SEED_PETUGAS_PASSWORD=$SEED_PETUGAS_PASSWORD
SEED_RESET_STAFF_PASSWORD=$SEED_RESET_STAFF_PASSWORD
ENVFILE
  chmod 600 .env
  ENV_CREATED="true"
  ok ".env dibuat dengan password dan secret acak"
else
  # Jangan menimpa secret yang sudah ada. Hanya samakan URL publik.
  set_env_key() {
    local key="$1" value="$2" escaped
    escaped="$(printf '%s' "$value" | sed -e 's/[&|\\]/\\&/g')"
    if grep -q "^${key}=" .env; then
      sed -i "s|^${key}=.*|${key}=${escaped}|" .env
    else
      printf '%s=%s\n' "$key" "$value" >> .env
    fi
  }
  set_env_key "FRONTEND_URL" "$PUBLIC_BASE_URL"
  set_env_key "BACKEND_PUBLIC_URL" "$PUBLIC_BASE_URL"
  set_env_key "NEXT_PUBLIC_API_BASE_URL" "$PUBLIC_BASE_URL/api/v1"
  set_env_key "BIND_HOST" "$STACK_BIND_HOST"
  set_env_key "APP_PORT" "$APP_PORT"
  ok ".env sudah ada; secret lama dipertahankan, URL publik diperbarui"
fi

if grep -Eq '^APP_ENV=(production|staging)' .env 2>/dev/null; then
  for key in JWT_ACCESS_SECRET JWT_REFRESH_SECRET MIDTRANS_SERVER_KEY DB_PASSWORD; do
    if ! grep -Eq "^${key}=.+" .env; then
      die "APP_ENV $(grep -E '^APP_ENV=' .env | head -1 | cut -d= -f2-) membutuhkan ${key} yang terisi di .env."
    fi
  done
  if grep -q '^APP_ENV=production' .env; then
    if ! grep -q '^MIDTRANS_IS_PRODUCTION=true' .env; then
      die "APP_ENV=production membutuhkan MIDTRANS_IS_PRODUCTION=true."
    fi
    if grep -Eq '^MIDTRANS_SERVER_KEY=SB-Mid-server' .env; then
      die "APP_ENV=production tetapi MIDTRANS_SERVER_KEY masih key sandbox (SB-Mid-server...). Gunakan key Mid-server... atau ubah APP_ENV menjadi local."
    fi
  fi
fi

# =============================================================================
# 7. Build dan jalankan stack
# =============================================================================
log "6/8 Build dan start container (bisa 5-20 menit pada build pertama)"
docker compose up -d --build

# =============================================================================
# 8. Tunggu aplikasi siap
# =============================================================================
log "7/8 Menunggu backend siap"
BACKEND_READY="false"
for _ in $(seq 1 80); do
  if curl -fsS --max-time 2 "http://127.0.0.1:$APP_PORT/healthz" >/dev/null 2>&1; then
    BACKEND_READY="true"
    break
  fi
  sleep 3
done
if [[ "$BACKEND_READY" != "true" ]]; then
  docker compose logs --tail=120 backend || true
  die "Backend tidak siap. Periksa log di atas."
fi
ok "Backend sehat"

log "Menunggu frontend siap"
FRONTEND_READY="false"
for _ in $(seq 1 60); do
  if curl -fsS --max-time 2 -o /dev/null "http://127.0.0.1:$APP_PORT/"; then
    FRONTEND_READY="true"
    break
  fi
  sleep 3
done
if [[ "$FRONTEND_READY" != "true" ]]; then
  docker compose logs --tail=120 frontend || true
  die "Frontend tidak siap. Periksa log di atas."
fi
ok "Frontend sehat"

# Firewall (opsional). Pada mode proxy, port aplikasi hanya listen di
# 127.0.0.1. Pada mode langsung, hanya APP_PORT yang dibuka.
if [[ "$ENABLE_UFW" == "true" ]] && command -v ufw >/dev/null 2>&1; then
  ufw allow 22/tcp >/dev/null 2>&1 || true
  if [[ "$EDGE_PROXY" == "true" ]]; then
    ufw allow 80/tcp >/dev/null 2>&1 || true
    ufw allow 443/tcp >/dev/null 2>&1 || true
  else
    ufw allow "$APP_PORT"/tcp >/dev/null 2>&1 || true
  fi
  ufw --force enable >/dev/null 2>&1 || true
fi

# =============================================================================
# Ringkasan
# =============================================================================
log "8/8 Selesai"
printf '%s\n' "============================================================"
printf 'URL aplikasi : %s\n' "$PUBLIC_BASE_URL"
printf 'Health check : %s/healthz\n' "$PUBLIC_BASE_URL"
printf 'Login staf   : %s/login\n' "$PUBLIC_BASE_URL"
printf 'Superadmin   : superadmin@kawahputih.id\n'
if [[ "$ENV_CREATED" == "true" ]]; then
  printf 'Password     : %s\n' "$SEED_SUPERADMIN_PASSWORD"
  printf '\nPassword staf lain dan semua secret tersimpan di: %s/.env\n' "$PROJECT_DIR"
else
  printf 'Password     : lihat SEED_SUPERADMIN_PASSWORD di %s/.env\n' "$PROJECT_DIR"
fi
if [[ -n "$OLD_BACKUP_FILE" ]]; then
  printf 'Backup DB lama: %s\n' "$OLD_BACKUP_FILE"
  printf 'Panduan impor ada di WORKSHOP.md (bagian "Migrasi data stack lama").\n'
fi
printf '%s\n' "============================================================"
printf '\nPerintah berguna:\n'
printf '  cd %s\n' "$PROJECT_DIR"
printf '  docker compose ps\n'
printf '  docker compose logs -f backend\n'
printf '  docker compose up -d --build     # setelah git pull / ubah kode\n'
if [[ -z "$DOMAIN" ]]; then
  printf '\nCatatan: aplikasi berjalan tanpa domain (lewat IP). Untuk memakai domain,\n'
  printf 'isi DOMAIN dan ENABLE_HTTPS=true di deploy.env lalu jalankan ulang skrip.\n'
fi
if [[ "$SCHEME" == "http" && -n "$DOMAIN" && "$EDGE_PROXY" == "true" ]]; then
  printf '\nCatatan: HTTPS belum aktif. Buka http://%s atau perbaiki DNS lalu jalankan ulang.\n' "$DOMAIN"
fi
if [[ "$EDGE_PROXY" != "true" ]]; then
  printf '\nCatatan: Nginx host tidak dipakai karena port 80 sudah terpakai.\n'
  printf 'Aplikasi diakses langsung di %s (pastikan port %s terbuka di firewall/cloud).\n' "$PUBLIC_BASE_URL" "$APP_PORT"
fi
