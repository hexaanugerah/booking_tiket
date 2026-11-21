# Panduan Workshop — Docker, VPS, dan CI/CD Tanpa Coolify

Dokumen ini menyesuaikan job sheet workshop Docker + Coolify + GitHub Actions
dengan projek Kawah Putih (Go + Next.js + MySQL), **tanpa perlu masuk ke
dashboard web/Coolify**.

Semua komponen (backend, frontend, database, Redis) dijalankan oleh Docker
Compose. Satu-satunya pintu masuk aplikasi adalah port **8080**.

---

## Tugas 1 — Containerization di lokal

### Jalankan

Windows:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\run-local.ps1
```

atau klik dua kali `scripts\run-local.cmd`.

Linux/macOS:

```bash
bash scripts/run-local.sh
```

Atau manual:

```bash
docker compose up -d --build
docker compose ps
```

Buka:

```text
http://localhost:8080
```

Jika `.env` belum ada, skrip akan membuatnya otomatis beserta akun demo:

```text
superadmin@kawahputih.id / superadmin123
admin@kawahputih.id      / admin123
manager@kawahputih.id    / manager123
petugas@kawahputih.id    / petugas123
```

Jika `.env` sudah ada, akun mengikuti nilai `SEED_*_PASSWORD` di file tersebut.

### Status verifikasi (21 November 2026)

Sudah diuji di Windows + Docker Desktop:

- `docker compose ps`: `mysql`, `redis`, `backend`, `frontend`, dan `gateway`
  berstatus `healthy`.
- `GET http://localhost:8080/healthz` mengembalikan `200`.
- Halaman utama, halaman login, dan API publik dapat diakses lewat satu port
  `8080`.
- Login `superadmin@kawahputih.id` berhasil dan mengembalikan token.

### Bukti Tugas 1

- Screenshot `docker compose ps` (semua service `running`/`healthy`).
- Screenshot browser di `http://localhost:8080`.
- Build manual image (opsional, untuk validasi):

```bash
docker build --target backend  -t kawahputih-backend:v1 .
docker build --target frontend -t kawahputih-frontend:v1 .
```

Catatan: projek ini full-stack (Go + Next.js + MySQL), jadi Tugas 1 memakai
Docker Compose, bukan satu container tunggal. Entry point tetap satu port,
yaitu `localhost:8080`, sama seperti ketentuan job sheet.

---

## Tugas 2 — Deploy ke VPS tanpa dashboard Coolify

### Di server (Ubuntu/Debian)

```bash
git clone https://github.com/hexaanugerah/kawahputihukk.git /opt/kawahputihukk
cd /opt/kawahputihukk/kawahputih
sudo bash scripts/deploy-server.sh
```

Skrip akan:

1. Install Docker, Docker Compose, dan alat dasar.
2. Memasang Nginx + Certbot hanya bila port 80 bebas dan HTTPS diminta.
3. Menyiapkan database MySQL dan Redis sebagai container.
4. Membuat `.env` beserta password/secret acak.
5. Build backend dan frontend.
6. Menjalankan migrasi + seed otomatis saat backend start.
7. Mengatur firewall dan menunggu aplikasi benar-benar sehat.

Setelah selesai, skrip menampilkan URL aplikasi dan password superadmin.

Tanpa domain, aplikasi bisa dibuka lewat URL yang dicetak skrip:

```text
http://IP_SERVER
```

atau, bila port 80 sudah dipakai aplikasi lain:

```text
http://IP_SERVER:8080
```

Skrip memakai `EDGE_PROXY=auto`: bila port 80 sudah dipakai Coolify/Traefik,
Nginx host dilewati dan gateway langsung dibuka di `APP_PORT`. Coolify tidak
perlu dimatikan dan tidak ada dashboard web yang perlu dibuka.

### Database

Database berjalan sebagai container `mysql:8.4` (LTS) di network internal Docker.
Database tidak diekspos ke internet. Backend otomatis menjalankan migrasi
GORM dan seed saat pertama kali start.

Akses database dari server (mis. untuk cek tabel):

```bash
docker compose exec mysql \
  mysql -u"$DB_USER" -p"$DB_PASSWORD" "$DB_NAME"
```

Atau dari komputer lokal lewat SSH tunnel (tanpa membuka port 3306):

```bash
ssh -L 3306:127.0.0.1:3306 user@IP_SERVER
```

### Migrasi data stack lama

Jika server sudah menjalankan `docker/docker-compose.prod.yml`, skrip deploy
otomatis membuat dump database ke `backups/db-sebelum-deploy-*.sql` sebelum
menghentikan stack lama.

Setelah stack baru sehat, impor dump tersebut **hanya bila data lama memang
ingin dipakai** (perintah ini menimpa tabel dengan nama sama):

```bash
docker compose stop backend
docker compose exec -T mysql sh -c \
  'exec mysql -uroot -p"$MYSQL_ROOT_PASSWORD" "$MYSQL_DATABASE"' \
  < backups/db-sebelum-deploy-XXXXXXXX.sql
docker compose start backend
```

---

## Tugas 2B — Domain dan HTTPS (opsional)

Jika domain sudah diarahkan ke server dan port 80/443 terbuka:

```bash
cp deploy.env.example deploy.env
nano deploy.env
```

Isi:

```env
DOMAIN=server.smkbudibakti.sch.id
ENABLE_HTTPS=true
LETSENCRYPT_EMAIL=email-anda@sekolah.sch.id
```

Lalu jalankan ulang:

```bash
sudo bash scripts/deploy-server.sh
```

Skrip menerbitkan sertifikat Let's Encrypt dan mengarahkan HTTP ke HTTPS.
Jika DNS belum mengarah ke server, skrip tetap menyelesaikan deployment lewat
HTTP dan memberi peringatan.

Jika tidak punya akses ke domain tersebut, cukup lewati bagian ini dan pakai
IP server. Aplikasi tetap berjalan lengkap.

---

## Tugas 3 — CI/CD GitHub Actions

Workflow ada di `.github/workflows/deploy.yml`.

### Secrets yang perlu diisi

Buka GitHub repository > Settings > Secrets and variables > Actions:

| Secret | Keterangan |
|---|---|
| `VPS_HOST` | IP/host server |
| `VPS_USERNAME` | user SSH (masukkan ke grup `docker`) |
| `VPS_SSH_KEY` | private key SSH |
| `VPS_PORT` | opsional, default 22 |
| `VPS_APP_DIR` | opsional, default `/opt/kawahputihukk/kawahputih` |

### Alur

1. Push ke branch `main`.
2. Job `build` memvalidasi `docker compose config` dan membangun image backend
   serta frontend.
3. Job `deploy` masuk ke VPS lewat SSH, menjalankan `git pull`, lalu
   `docker compose up -d --build`.
4. Workflow mengecek `http://127.0.0.1:8080/healthz` sampai aplikasi sehat.

Tidak ada webhook Coolify dan tidak ada login dashboard.

---

## Ringkasan checklist job sheet

- [x] Dockerfile dan `.dockerignore` ada di root.
- [x] Aplikasi berjalan di `localhost:8080` lewat Docker Compose.
- [x] Database MySQL aktif sebagai container dan terhubung ke backend.
- [x] Migrasi dan seed berjalan otomatis saat backend start.
- [x] Deploy VPS tanpa dashboard web (Docker Compose + Nginx).
- [x] Domain/HTTPS opsional lewat `deploy.env` + Certbot.
- [x] Workflow CI/CD GitHub Actions untuk build dan deploy via SSH.

---

## Perintah operasional

```bash
docker compose ps
docker compose logs -f backend
docker compose logs -f frontend
docker compose logs -f mysql
docker compose restart backend
docker compose up -d --build
docker compose down
```

`docker compose down` tidak menghapus data. Jangan gunakan `docker compose down -v`
kecuali memang ingin menghapus database dan file upload.
