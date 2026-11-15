# Deploy Kawah Putih ke Railway dan Netlify

Panduan ini memakai Railway untuk backend Go, TiDB Cloud Starter untuk MySQL-compatible, dan Netlify untuk Next.js. Redis tidak perlu dibuat; backend melewatinya jika `REDIS_ADDR` tidak diisi.

## 1. Buat database TiDB

1. Buat satu cluster **TiDB Cloud Starter**.
2. Dari halaman cluster, buka **Connect** dan salin host, port, username, password, serta nama database.
3. Koneksi dari aplikasi wajib memakai TLS. Backend sudah menyediakan `DB_TLS=true` untuk ini.

Simpan kredensial dengan aman. Jangan commit kredensial ke GitHub.

## 2. Deploy backend di Railway

1. Push perubahan proyek ini ke GitHub.
2. Di Railway, buat project lalu pilih **Deploy from GitHub Repo** dan pilih repositori ini.
3. Buka service backend → **Settings** → **Build** dan atur **Root Directory** menjadi `/backend`. Railway akan menemukan `backend/Dockerfile` dan membangun image backend.
4. Pastikan public networking aktif untuk service backend. Aplikasi membaca port Railway dari `PORT` secara otomatis.
5. Buka **Variables** dan tambahkan:

   ```text
   APP_ENV=staging
   BACKEND_PUBLIC_URL=https://<domain-backend-railway>
   FRONTEND_URL=https://<domain-situs-netlify>
   DB_HOST=<host-dari-TiDB>
   DB_PORT=<port-dari-TiDB>
   DB_USER=<username-dari-TiDB>
   DB_PASSWORD=<password-dari-TiDB>
   DB_NAME=<nama-database-dari-TiDB>
   DB_TLS=true
   DB_MAX_OPEN_CONNS=10
   DB_MAX_IDLE_CONNS=5
   UPLOAD_DIR=/app/uploads
   JWT_ACCESS_SECRET=<secret-acak-panjang-1>
   JWT_REFRESH_SECRET=<secret-acak-panjang-2>
   MIDTRANS_SERVER_KEY=<server-key-sandbox-Midtrans>
   MIDTRANS_IS_PRODUCTION=false
   ```

   Untuk deploy pertama, gunakan URL placeholder `FRONTEND_URL` sementara. Setelah Netlify memberikan URL, ganti nilainya dan deploy ulang backend.

6. Deploy service. Di **Settings → Networking**, buat domain publik jika Railway belum membuatkannya. Gunakan URL HTTPS domain itu untuk `BACKEND_PUBLIC_URL`.
7. Buka `https://<domain-backend-railway>/healthz`; hasil yang diharapkan adalah JSON `{"status":"ok"}`.

`APP_ENV=staging` diperlukan karena konfigurasi backend meminta JWT secrets dan Midtrans key pada lingkungan staging/production. Gunakan server key sandbox, bukan production.

## 4. Deploy frontend di Netlify

1. Di Netlify pilih **Add new site → Import an existing project**, hubungkan GitHub, lalu pilih repositori ini.
2. Konfigurasi di root `netlify.toml` memakai folder `frontend`, perintah `npm run build`, dan output `.next`. Netlify mendukung Next.js dengan runtime-nya; bila form setup meminta nilainya, masukkan nilai tersebut.
3. Pada **Site configuration → Environment variables**, tambahkan. Pastikan kedua variabel aktif untuk Production; aktifkan juga untuk Deploy Previews jika preview branch perlu memanggil API:

   ```text
   NEXT_PUBLIC_API_BASE_URL=https://<domain-backend-railway>/api/v1
   BACKEND_ORIGIN=https://<domain-backend-railway>
   ```

   Jangan beri garis miring di akhir `BACKEND_ORIGIN`. `NEXT_PUBLIC_API_BASE_URL` dipakai browser untuk API dan `BACKEND_ORIGIN` untuk rewrite Next.js, termasuk berkas `/uploads`.

4. Deploy site dan salin URL HTTPS Netlify, misalnya `https://nama-situs.netlify.app`.
5. Kembali ke Railway, isi `FRONTEND_URL` dengan URL tersebut, lalu redeploy backend agar CORS mengizinkan frontend.

## 5. Uji aplikasi

1. Buka URL Netlify.
2. Coba halaman publik, daftar/login, dan fitur yang membaca data dari API.
3. Bila API gagal, pastikan URL backend di kedua environment variables Netlify sama dengan domain Railway.
4. Bila muncul error CORS, pastikan `FRONTEND_URL` di Railway sama persis dengan origin Netlify (tanpa path).
5. Bila koneksi database gagal, periksa host/port/kredensial TiDB dan pastikan `DB_TLS=true`.

## Catatan penyimpanan upload dan biaya

Backend menyimpan berkas galeri dan bukti pembayaran pada disk lokal. Penyimpanan lokal Railway bersifat sementara kecuali service diberi Volume. Jika upload harus bertahan setelah redeploy, buat Volume dan mount ke `/app/uploads`; biaya resource/penyimpanan tetap mengikuti kuota akun Railway.

Railway Trial memberi kredit $5 untuk 30 hari, lalu Free memberi kredit resource $1 per bulan. Monitor pemakaian dari dashboard Railway. Netlify dan TiDB juga mempunyai kuota paket gratis masing-masing.
