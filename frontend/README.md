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

## Deploy ke Cloudflare Workers

Frontend di-deploy sebagai Cloudflare Worker lewat adapter OpenNext
(`open-next.config.ts` + `wrangler.jsonc`).

```bash
npx wrangler login        # sekali saja
npm run cf:build          # next build + bundle Worker
npm run cf:deploy         # upload ke Cloudflare
npm run cf:preview        # uji hasil build di lokal (port 8787)
```

Catatan penting:

1. `NEXT_PUBLIC_API_BASE_URL` di-*inline* saat build, jadi isi
   `.env.production.local` dengan URL backend yang bisa diakses publik
   (HTTPS) **sebelum** `npm run cf:build`.
2. Backend Go (MySQL, Redis, penyimpanan `uploads/`) tidak bisa dijalankan
   di Workers — host terpisah (VPS/Docker). Daftarkan domain Worker di
   `FRONTEND_URL` backend (daftar dipisah koma) supaya CORS dan redirect
   OAuth Google mengarah ke domain yang benar.
3. Nama Worker diatur di `wrangler.jsonc` (`name: "kawahputih"`).
4. Worker meneruskan `/api/v1/*` dan `/uploads/*` ke `BACKEND_ORIGIN`
   (default `http://localhost:8080`). Karena Worker tidak bisa menjangkau LAN,
   isi `BACKEND_ORIGIN` dengan URL backend yang publik saat build:

   ```bash
   BACKEND_ORIGIN=https://<url-backend-publik> npx opennextjs-cloudflare build
   CI=true npx wrangler deploy
   ```

   Kalau `BACKEND_ORIGIN` tidak diisi, situs hasil deploy akan memanggil
   `localhost` dan API-nya gagal. Proxy ini juga membuat CORS tidak relevan
   karena browser hanya bicara ke origin Worker.
4. `cf:build` memindahkan isi `.next` ke `.open-next`, sehingga `next start`
   lokal akan gagal / menyajikan halaman tanpa CSS setelahnya. Untuk lokal
   pakai `npm run dev`, atau jalankan `npm run build` lagi sebelum
   `npm run start`.
