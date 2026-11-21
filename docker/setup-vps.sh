#!/usr/bin/env bash
# =============================================================================
# DEPRECATED: skrip setup lama untuk docker/docker-compose.prod.yml.
#
# Sekarang seluruh stack (frontend + backend + MySQL + Redis + gateway)
# dideploy lewat skrip di root projek. Skrip ini dipertahankan agar
# pemanggilan lama tetap bekerja, tetapi isinya hanya meneruskan ke:
#   scripts/deploy-server.sh
#
# Cara yang disarankan di server baru:
#   curl -fsSL https://raw.githubusercontent.com/hexaanugerah/booking_tiket/main/scripts/bootstrap.sh | sudo bash
# =============================================================================
set -Eeuo pipefail

REPO_URL="${REPO_URL:-https://github.com/hexaanugerah/booking_tiket.git}"
APP_DIR="${APP_DIR:-/opt/booking_tiket}"

if [[ "${EUID}" -ne 0 ]]; then
  echo "[x] Jalankan sebagai root: sudo bash docker/setup-vps.sh" >&2
  exit 1
fi

export DEBIAN_FRONTEND=noninteractive
if ! command -v git >/dev/null 2>&1; then
  apt-get update -y
  apt-get install -y ca-certificates curl git
fi

if [[ -d "$APP_DIR/.git" ]]; then
  echo "[i] Memperbarui kode di $APP_DIR"
  git -c safe.directory="$APP_DIR" -C "$APP_DIR" pull --ff-only
elif [[ -e "$APP_DIR" ]]; then
  echo "[x] $APP_DIR sudah ada tetapi bukan repo git." >&2
  exit 1
else
  echo "[i] Meng-clone $REPO_URL ke $APP_DIR"
  mkdir -p "$(dirname "$APP_DIR")"
  git clone "$REPO_URL" "$APP_DIR"
fi

cd "$APP_DIR"
exec bash scripts/deploy-server.sh
