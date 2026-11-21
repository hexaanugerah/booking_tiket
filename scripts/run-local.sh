#!/usr/bin/env bash
# Jalankan dari root projek (Linux/macOS):
#   bash scripts/run-local.sh
#
# Tidak perlu domain, DNS, Nginx, atau panel web. Semua berjalan lokal.
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")/.."

if ! command -v docker >/dev/null 2>&1; then
  echo "Docker belum terpasang." >&2
  exit 1
fi
if ! docker info >/dev/null 2>&1; then
  echo "Docker daemon belum berjalan." >&2
  exit 1
fi
if ! docker compose version >/dev/null 2>&1; then
  echo "Docker Compose plugin belum tersedia." >&2
  exit 1
fi

ENV_CREATED=false
if [[ ! -f .env ]]; then
  echo "Membuat .env lokal dengan akun demo..."
  cat > .env <<'EOF'
# Dibuat otomatis oleh scripts/run-local.sh
APP_ENV=local
FRONTEND_URL=http://localhost:8080,http://127.0.0.1:8080
BACKEND_PUBLIC_URL=http://localhost:8080
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080/api/v1
BIND_HOST=127.0.0.1
APP_PORT=8080

SEED_SUPERADMIN_PASSWORD=superadmin123
SEED_ADMIN_PASSWORD=admin123
SEED_MANAGER_PASSWORD=manager123
SEED_PETUGAS_PASSWORD=petugas123
EOF
  if command -v openssl >/dev/null 2>&1; then
    {
      printf 'JWT_ACCESS_SECRET=%s\n' "$(openssl rand -hex 32)"
      printf 'JWT_REFRESH_SECRET=%s\n' "$(openssl rand -hex 32)"
    } >> .env
  fi
  chmod 600 .env
  ENV_CREATED=true
fi

echo "Membangun dan menjalankan container..."
docker compose up -d --build

echo "Menunggu backend siap..."
backend_ready=false
for _ in $(seq 1 80); do
  if curl -fsS --max-time 2 http://127.0.0.1:8080/healthz >/dev/null 2>&1; then
    backend_ready=true
    break
  fi
  sleep 3
done
if [[ "$backend_ready" != "true" ]]; then
  docker compose logs --tail=120 backend
  echo "Backend tidak siap." >&2
  exit 1
fi

echo "Menunggu frontend siap..."
frontend_ready=false
for _ in $(seq 1 60); do
  if curl -fsS --max-time 2 -o /dev/null http://127.0.0.1:8080; then
    frontend_ready=true
    break
  fi
  sleep 3
done
if [[ "$frontend_ready" != "true" ]]; then
  docker compose logs --tail=120 frontend
  echo "Frontend tidak siap." >&2
  exit 1
fi

echo ""
echo "Aplikasi siap dipakai:"
echo "  Aplikasi : http://localhost:8080"
echo "  API      : http://localhost:8080/api/v1"
echo "  Login    : http://localhost:8080/login"
if [[ "$ENV_CREATED" == "true" ]]; then
  echo "  Superadmin: superadmin@kawahputih.id / superadmin123"
  echo "  Admin     : admin@kawahputih.id / admin123"
fi
echo ""

if command -v xdg-open >/dev/null 2>&1; then
  xdg-open "http://localhost:8080" >/dev/null 2>&1 || true
elif command -v open >/dev/null 2>&1; then
  open "http://localhost:8080" >/dev/null 2>&1 || true
fi
