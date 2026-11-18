# Frontend — Kawah Putih Rancabali (Part 2.3)

## Status jujur

Struktur folder mengikuti Part 2.3 secara lengkap. **Halaman fungsional
cuma dibangun untuk 7 domain yang backend-nya benar-benar ada** (auth,
user/staff, article, gallery, package, booking, dan sedikit role lewat
admin users). Sisa ~20 route yang didaftar dokumen (destinasi, kategori,
fasilitas, event, FAQ, banner, promo, voucher, kontak, laporan, analytics,
notifikasi, audit log, dst) dibuat sebagai **halaman "Coming Soon"** yang
jelas menyatakan dirinya belum tersambung ke backend — bukan halaman
kosong atau data palsu.

## Belum pernah di-`npm install` / `npm run build`

Sama seperti backend, saya tidak punya akses ke registry npm (`registry.npmjs.org`
sebenarnya ada di allowlist jaringan saya, jadi ini kemungkinan besar
bisa jalan — tapi saya belum sempat coba langsung). Jalankan:

```bash
cd frontend
npm install
npm run dev
```

dan laporkan kalau ada error TypeScript — kemungkinan ada 1-2 tipe yang
meleset di tempat yang tidak saya periksa ulang.

## Keputusan desain yang perlu kamu tahu

1. **Routing paket & artikel pakai ID, bukan slug** — backend hasil
   rebuild Part 2.2 cuma punya `FindByID` (bukan `FindBySlug` seperti versi
   sebelumnya). URL jadi `/packages/<uuid>` bukan `/packages/kawah-putih-basic`.
   Kurang SEO-friendly, tapi jujur sesuai kapasitas backend saat ini. Kalau mau
   URL slug, backend perlu ditambah endpoint `FindBySlug` dulu.

2. **QR Ticket di-generate di frontend** (`react-qr-code`), backend cuma
   kirim `ticket_code` — sesuai keputusan desain sejak Fase 2 backend.

3. **`store/notification.store.ts`, `wishlist.store.ts`, `language.store.ts`
   sengaja inert** (cuma nyimpen state lokal) — ditandai jelas di komentar
   kode karena tidak ada backend module yang mendukungnya (Notification,
   Wishlist server-side, i18n).

4. **Halaman profil read-only** — backend tidak punya endpoint update profil
   (FR-006 di BRD belum diimplementasi backend-nya).

5. **Check-in tiket manual (ketik kode)**, bukan scan kamera langsung —
   scanner QR fisik (kamera browser) belum diintegrasikan, cuma input teks.

6. **Role check di `middleware.ts` cuma cek keberadaan cookie**, bukan
   decode isi JWT — dijelaskan di komentar kenapa (menghindari duplikasi
   logic verifikasi signature di dua tempat). Guard role sesungguhnya
   tetap di backend (`RequireRole` middleware) + client-side check di
   `lib/permission.ts` untuk UX saja.

## Environment variable

Buat `.env.local` di folder `frontend/`:
```
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080/api/v1
```

## Buka dari HP atau PC lain (satu jaringan)

1. Cari IPv4 komputer ini: `ipconfig` (mis. `192.168.1.5`).
2. Jalankan backend (port 8080) dan frontend. `npm run dev` sudah bind ke
   `0.0.0.0`, jadi bisa diakses dari perangkat lain.
3. Di HP/PC lain buka `http://192.168.1.5:3000` — atau port yang dicetak Next
   (mis. `3001` kalau 3000 sedang dipakai). Jangan pakai `localhost` dari HP,
   karena itu menunjuk ke HP itu sendiri.
4. `lib/api-base.ts` otomatis mengganti host `localhost` pada
   `NEXT_PUBLIC_API_BASE_URL` dengan hostname halaman yang dibuka, jadi API
   tetap mengarah ke `http://192.168.1.5:8080`. Backend (APP_ENV selain
   `production`) mengizinkan origin LAN seperti `http://192.168.1.5:3001`.
5. Kalau perangkat lain tidak bisa connect, izinkan port di Windows Firewall
   (jalankan sebagai Administrator):

   ```bash
   netsh advfirewall firewall add rule name="KawahPutih dev" dir=in action=allow protocol=TCP localport=3000-3001,8080
   ```

## Buka dari luar rumah (internet)

Jalankan:

```bash
bash scripts/start-public.sh
```

Skrip menyalakan backend dan frontend bila belum jalan, membuat satu quick
tunnel Cloudflare ke frontend (port 3000), lalu mencetak URL
`https://<acak>.trycloudflare.com` yang bisa dibuka dari jaringan mana pun.
Server Next meneruskan `/api/v1/*` dan `/uploads/*` ke backend, jadi cukup satu
tunnel: tidak perlu port forwarding, buka firewall, atau build/deploy Worker.

Catatan:

1. Hanya koneksi keluar yang dipakai, jadi router tidak perlu diatur.
2. MySQL dan Redis harus sudah jalan (backend butuh keduanya).
3. URL quick tunnel berubah setiap skrip dijalankan.
4. Login Google OAuth butuh URL callback tetap; dengan URL acak, atur
   `FRONTEND_URL` backend ke URL publik tersebut agar redirect-nya benar.
   Untuk URL permanen, buat named tunnel Cloudflare (butuh akun + domain) ke
   `http://localhost:3000`.

## Deploy ke Vercel (frontend) + Railway (backend)

Konfigurasi yang dipakai saat ini: halaman di Vercel, API Go di Railway.

1. Vercel → Project Settings → Environment Variables:
   `NEXT_PUBLIC_API_BASE_URL=https://<backend>.up.railway.app/api/v1`.
   Nilai `NEXT_PUBLIC_*` di-*inline* saat build, jadi setelah mengubahnya
   lakukan redeploy (build baru).
2. Railway (service backend) → `FRONTEND_URL=https://<frontend>.vercel.app`.
   Variabel ini dipakai untuk CORS dan redirect OAuth; beberapa origin boleh
   dipisah koma.
3. `BACKEND_ORIGIN` hanya dibutuhkan bila memakai proxy runtime
   (`app/api/v1/[...path]/route.ts` + `app/uploads/[...path]/route.ts`).
   Dengan `NEXT_PUBLIC_API_BASE_URL` absolut ke Railway, browser memanggil
   Railway langsung dan CORS backend yang menanganinya.
4. Bukti pembayaran disimpan di `UPLOAD_DIR` backend. Di Railway pasang volume
   permanen (mount path `/app/uploads`, `UPLOAD_DIR=/app/uploads`) agar berkas
   tidak hilang saat container restart.

## Deploy ke Cloudflare Workers

Frontend di-deploy sebagai Cloudflare Worker lewat adapter OpenNext
(`open-next.config.ts` + `wrangler.jsonc`).

```bash
npx wrangler login        # sekali saja
npm run cf:build          # next build + bundle Worker
npm run cf:deploy         # upload ke Cloudflare
npm run cf:preview        # uji hasil build di lokal (port 8787)
```

### 1. Siapkan backend dengan URL publik yang tetap

Backend Go (MySQL, Redis, `uploads/`) tidak bisa berjalan di Workers. Worker
juga **tidak bisa menjangkau `localhost` atau IP LAN**, jadi backend harus
punya URL HTTPS yang bisa diakses dari internet, misalnya:

- VPS + domain: `https://api.kawahputih.id`
- Cloudflare Named Tunnel: `https://api.kawahputih.id`
- Platform lain: `https://<service>.onrender.com`, dsb.

Quick Tunnel (`https://<acak>.trycloudflare.com`) berubah setiap restart dan
tidak cocok untuk konfigurasi tetap. Tes dulu dari internet:

```bash
curl https://api.kawahputih.id/healthz
```

### 2. Atur `BACKEND_ORIGIN` (runtime, tanpa build ulang)

Proxy `/api/v1/*` dan `/uploads/*` sekarang ditangani route handler runtime
(`app/api/v1/[...path]/route.ts` + `app/uploads/[...path]/route.ts`). Nilai
`BACKEND_ORIGIN` dibaca pada setiap request, bukan dibekukan saat build.

Isi URL backend (tanpa `/api/v1`) di `wrangler.jsonc`:

```jsonc
"vars": {
  "BACKEND_ORIGIN": "https://api.kawahputih.id"
}
```

Alternatif tanpa mengubah file:

```bash
npx wrangler secret put BACKEND_ORIGIN
```

Setelah build pertama, mengubah URL backend cukup dengan mengubah nilai itu
lalu `npm run cf:deploy` — tidak perlu `cf:build` lagi. Kalau nilainya kosong,
API mengembalikan HTTP 503 dengan pesan `BACKEND_ORIGIN belum diatur`.

### 3. Catatan lain

1. `NEXT_PUBLIC_API_BASE_URL` masih di-*inline* saat build, tetapi
   `lib/api-base.ts` mengubahnya menjadi path relatif `/api/v1` ketika halaman
   dibuka dari domain publik. Jadi build LAN yang sudah ada tetap bisa dipakai
   di Worker selama `BACKEND_ORIGIN` diisi.
2. Daftarkan domain Worker di `FRONTEND_URL` backend (dipisah koma) untuk
   redirect OAuth Google. Untuk API lewat proxy, CORS tidak relevan karena
   browser hanya bicara ke origin Worker.
3. Nama Worker diatur di `wrangler.jsonc` (`name: "kawahputih"`).
4. `cf:build` memindahkan isi `.next` ke `.open-next`, sehingga `next start`
   lokal akan gagal / menyajikan halaman tanpa CSS setelahnya. Untuk lokal
   pakai `npm run dev`, atau jalankan `npm run build` lagi sebelum
   `npm run start`.

