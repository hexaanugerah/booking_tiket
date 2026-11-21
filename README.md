# Kawah Putih

Aplikasi full-stack: backend Go, frontend Next.js, database MySQL, Redis, dan
gateway Nginx. Seluruh stack berjalan dengan Docker Compose; database tidak
dibuka ke internet.

Panduan workshop lengkap ada di [`WORKSHOP.md`](WORKSHOP.md).

---

## Jalankan di komputer lokal

Windows (klik dua kali atau jalankan):

```powershell
scripts\run-local.cmd
```

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

Semua service (MySQL, Redis, backend, frontend, gateway) berjalan sekaligus.
Saat pertama kali dijalankan, `.env` dibuat otomatis beserta akun demo:

```text
superadmin@kawahputih.id / superadmin123
admin@kawahputih.id      / admin123
manager@kawahputih.id    / manager123
petugas@kawahputih.id    / petugas123
```

Untuk membuka dari HP di jaringan Wi-Fi yang sama, ubah `.env`:

```env
BIND_HOST=0.0.0.0
```

Lalu jalankan ulang `docker compose up -d` dan buka `http://IP_KOMPUTER:8080`
dari HP.

Hentikan tanpa menghapus data:

```bash
docker compose down
```

---

## Deploy ke server (tanpa dashboard web/Coolify)

Di server Ubuntu/Debian:

```bash
git clone https://github.com/hexaanugerah/kawahputihukk.git /opt/kawahputihukk
cd /opt/kawahputihukk/kawahputih
sudo bash scripts/deploy-server.sh
```

Skrip akan memasang Docker, membuat `.env` dengan password/secret acak,
menjalankan MySQL + Redis + backend + frontend + gateway, menjalankan migrasi
dan seed otomatis, lalu menampilkan URL aplikasi dan password superadmin.

- Jika port 80 bebas, Nginx host dipakai dan aplikasi dibuka lewat
  `http://IP_SERVER`.
- Jika port 80 sudah dipakai Coolify/Traefik, skrip otomatis melewati Nginx
  host dan membuka aplikasi di `http://IP_SERVER:8080`. Coolify tidak perlu
  dimatikan dan tidak ada dashboard yang perlu dibuka.

### Domain + HTTPS (opsional)

```bash
cp deploy.env.example deploy.env
nano deploy.env
```

Isi minimal:

```env
DOMAIN=server.smkbudibakti.sch.id
ENABLE_HTTPS=true
LETSENCRYPT_EMAIL=email-anda@sekolah.sch.id
```

Jalankan ulang:

```bash
sudo bash scripts/deploy-server.sh
```

Jika DNS belum mengarah ke server atau port 80 tertutup, skrip tetap
menjalankan aplikasi lewat HTTP dan memberi peringatan.

---

## CI/CD GitHub Actions

Workflow: [`.github/workflows/deploy.yml`](.github/workflows/deploy.yml).

Isi secrets di GitHub > Settings > Secrets and variables > Actions:

| Secret | Keterangan |
|---|---|
| `VPS_HOST` | IP/host server |
| `VPS_USERNAME` | user SSH yang masuk grup `docker` |
| `VPS_SSH_KEY` | private key SSH |
| `VPS_PORT` | opsional, default 22 |
| `VPS_APP_DIR` | opsional, default `/opt/kawahputihukk/kawahputih` |

Jalankan `scripts/deploy-server.sh` sekali di server agar `.env` tersedia.
Setelah itu setiap push ke `main` akan build image lalu deploy lewat SSH.

---

## Port dan data

| Service | Port host | Akses |
|---|---|---|
| Gateway (satu-satunya pintu aplikasi) | `8080` | publik |
| MySQL | tidak dipublikasikan | internal Docker |
| Redis | tidak dipublikasikan | internal Docker |
| Backend | tidak dipublikasikan | lewat gateway |
| Frontend | tidak dipublikasikan | lewat gateway |

Data disimpan di volume Docker `mysql-data` dan `uploads`. Perintah
`docker compose down` tidak menghapus data; hindari `docker compose down -v`
kecuali memang ingin menghapus database dan file upload.
