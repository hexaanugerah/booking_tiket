#!/usr/bin/env bash
# =============================================================================
# Bootstrap deploy Kawah Putih di server Ubuntu/Debian.
#
# Cukup satu perintah di server (sebagai root):
#   curl -fsSL https://raw.githubusercontent.com/hexaanugerah/booking_tiket/main/scripts/bootstrap.sh | sudo bash
#
# Skrip ini:
#   1. memasang git + curl bila belum ada,
#   2. clone repo (atau update bila sudah ada) ke APP_DIR,
#   3. menyerahkan sisanya ke scripts/deploy-server.sh (Docker, .env acak,
#      build, migrasi/seed, health check, dan URL publik).
#
# Override bila perlu:
#   APP_DIR=/opt/kawahputih REPO_URL=... BRANCH=main \
#     bash scripts/bootstrap.sh
# =============================================================================
set -Eeuo pipefail

REPO_URL="${REPO_URL:-https://github.com/hexaanugerah/booking_tiket.git}"
APP_DIR="${APP_DIR:-/opt/booking_tiket}"
BRANCH="${BRANCH:-main}"

if [[ "${EUID}" -ne 0 ]]; then
  echo "[x] Jalankan sebagai root: sudo bash scripts/bootstrap.sh" >&2
  exit 1
fi

export DEBIAN_FRONTEND=noninteractive
if ! command -v git >/dev/null 2>&1 || ! command -v curl >/dev/null 2>&1; then
  apt-get update -y
  apt-get install -y ca-certificates curl git
fi

if [[ -d "$APP_DIR/.git" ]]; then
  echo "[i] Memperbarui kode di $APP_DIR"
  git -c safe.directory="$APP_DIR" -C "$APP_DIR" fetch --prune origin
  git -c safe.directory="$APP_DIR" -C "$APP_DIR" checkout "$BRANCH"
  git -c safe.directory="$APP_DIR" -C "$APP_DIR" pull --ff-only origin "$BRANCH"
elif [[ -e "$APP_DIR" ]]; then
  echo "[x] $APP_DIR sudah ada tetapi bukan repo git." >&2
  echo "    Pindahkan/rename folder itu, atau set APP_DIR ke lokasi lain." >&2
  exit 1
else
  echo "[i] Meng-clone $REPO_URL ke $APP_DIR"
  mkdir -p "$(dirname "$APP_DIR")"
  git clone --branch "$BRANCH" "$REPO_URL" "$APP_DIR"
fi

cd "$APP_DIR"
exec bash scripts/deploy-server.sh
