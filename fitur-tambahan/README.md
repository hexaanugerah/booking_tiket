# Fitur Tambahan — Refund, Reschedule, dan Filter Ulasan

Folder ini **tidak mengganggu proyek utama**. Semua file berada di luar
`backend/` dan `frontend/`, sehingga tidak ikut `go build`, `next build`, atau
TypeScript checking.

## Isi folder

| Folder | Isi |
| --- | --- |
| `kosong/` | Kerangka file dengan isi **0 byte** (kosong). Path-nya sama persis dengan versi lengkap, jadi Anda bisa mengetik sendiri lalu menyalinnya ke proyek utama. |
| `contoh/` | Versi lengkap yang sudah terisi, sebagai referensi/copy-paste. |

Nama file di `kosong/` sengaja dibuat sama dengan `contoh/`, misalnya:

```
kosong/frontend/app/(admin)/admin/refunds/page.tsx
contoh/frontend/app/(admin)/admin/refunds/page.tsx
```

## Fitur yang tersedia di `contoh/`

1. **Refund**
   - Pengunjung mengajukan refund dari halaman Riwayat.
   - Admin melihat daftar, menyetujui/menolak, lalu menandai dana sudah dikirim.
   - Saat status menjadi `REFUNDED`, backend otomatis mengubah payment menjadi
     `REFUNDED`, booking menjadi `CANCELLED`, dan e-ticket menjadi `CANCELLED`.
2. **Reschedule**
   - Pengunjung mengubah tanggal kunjungan dari halaman Riwayat.
   - Maksimal 2 kali per booking.
   - Hanya untuk booking `PAID`/`CONFIRMED` dan diajukan sebelum tanggal kunjungan.
   - Kuota tiket pada tanggal baru ikut dicek agar tidak overbook.
   - Setiap perubahan dicatat di tabel `booking_reschedules`.
3. **Urutan & filter ulasan**
   - Halaman "Lihat semua ulasan" punya pilihan urutan:
     `Terbaru`, `Rating Tertinggi`, `Rating Terendah`.
   - Filter rating 1–5 (termasuk tombol `Semua`).
   - Filter dan urutan diproses di backend lewat query `rating` dan `sort`.

## Cara memakai

Pilih salah satu:

- **Ingin mengetik sendiri:** buka `kosong/`, ketik implementasi di file yang
  sudah disiapkan, lalu salin ke path yang sama di proyek utama.
- **Ingin langsung jalan:** salin isi `contoh/` ke path yang sama di proyek
  utama, lalu ikuti langkah integrasi di bawah.

> Jangan menyalin folder `fitur-tambahan/` ke dalam `frontend/app` atau
> `backend/internal`. Salin **file satu per satu** sesuai path-nya.

## Integrasi backend

### 1. Daftarkan model ke AutoMigrate

Di `backend/cmd/server/bootstrap.go`, tambahkan import:

```go
"github.com/KawahPutih/backend/internal/domain/refund"
"github.com/KawahPutih/backend/internal/domain/reschedule"
```

Lalu tambahkan model ke daftar `db.AutoMigrate(...)`:

```go
&refund.Refund{},
&reschedule.Reschedule{},
```

### 2. Daftarkan handler di bootstrap

Di `backend/internal/bootstrap/bootstrap.go`:

```go
// import
"github.com/KawahPutih/backend/internal/domain/refund"
"github.com/KawahPutih/backend/internal/domain/reschedule"
```

Tambahkan field di struct `App`:

```go
RefundHandler     *refund.Handler
RescheduleHandler *reschedule.Handler
```

Buat handler setelah `reviewHandler`:

```go
refundRepo := refund.NewRepository(db)
refundService := refund.NewService(refundRepo)
refundHandler := refund.NewHandler(refundService)

rescheduleRepo := reschedule.NewRepository(db)
rescheduleService := reschedule.NewService(rescheduleRepo)
rescheduleHandler := reschedule.NewHandler(rescheduleService)
```

Tambahkan ke nilai balik `&App{...}`:

```go
RefundHandler:     refundHandler,
RescheduleHandler: rescheduleHandler,
```

### 3. Daftarkan route

Di `backend/internal/routes/routes.go`, tambahkan import dan pemanggilan:

```go
import (
    "github.com/KawahPutih/backend/internal/domain/refund"
    "github.com/KawahPutih/backend/internal/domain/reschedule"
)

// setelah review.RegisterRoutes(...)
refund.RegisterRoutes(v1, app.RefundHandler, app.JWTManager)
reschedule.RegisterRoutes(v1, app.RescheduleHandler, app.JWTManager)
```

### 4. Ganti file review yang berubah

Salin 4 file ini dari `contoh/backend/internal/domain/review/`:

- `errors.go`
- `repository.go`
- `service.go`
- `handler.go`

Perubahan intinya: `ListApproved` sekarang menerima filter `rating` dan `sort`.

### 5. Migration SQL (opsional)

AutoMigrate sudah membuat tabel otomatis. File
`contoh/database/migrations/004_refund_reschedule.up.sql` hanya diperlukan bila
Anda ingin menyiapkan database lewat SQL manual.

## Integrasi frontend

Salin file berikut ke path yang sama:

```
contoh/frontend/types/refund.ts
contoh/frontend/types/reschedule.ts
contoh/frontend/services/refund.api.service.ts
contoh/frontend/services/reschedule.api.service.ts

contoh/frontend/features/refund/components/refund-request-dialog.tsx
contoh/frontend/features/booking/components/reschedule-dialog.tsx

contoh/frontend/app/(admin)/admin/refunds/page.tsx
contoh/frontend/app/(dashboard)/dashboard/history/page.tsx

contoh/frontend/features/public/components/public-review-list.tsx
contoh/frontend/services/review.api.service.ts
contoh/frontend/types/review.ts
```

Halaman `history/page.tsx` versi contoh sudah memuat tombol **Refund** dan
**Reschedule** sekaligus, jadi tidak perlu mengubah halaman lain.

## Ringkasan endpoint

| Method | Endpoint | Akses |
| --- | --- | --- |
| `POST` | `/api/v1/refunds` | Pengunjung login |
| `GET` | `/api/v1/refunds/mine` | Pengunjung login |
| `GET` | `/api/v1/admin/refunds` | Admin |
| `GET` | `/api/v1/admin/refunds/summary` | Admin |
| `PATCH` | `/api/v1/admin/refunds/:id/status` | Admin |
| `POST` | `/api/v1/bookings/:id/reschedule` | Pemilik booking |
| `GET` | `/api/v1/bookings/:id/reschedules` | Pemilik booking |
| `GET` | `/api/v1/reviews?rating=5&sort=rating_desc` | Publik |

Nilai `sort`: `newest`, `rating_desc`, `rating_asc`.

## Aturan bisnis singkat

| Fitur | Aturan |
| --- | --- |
| Refund | Hanya booking `PAID`/`CONFIRMED`; diajukan paling lambat H-1; nominal = total booking; satu pengajuan aktif per booking. |
| Refund admin | `PENDING → APPROVED/REJECTED`; `APPROVED → REFUNDED`. |
| Reschedule | Hanya booking `PAID`/`CONFIRMED`; sebelum tanggal kunjungan; tanggal baru minimal besok; maksimal 2 kali; kuota dicek. |
| Ulasan | Rating 1–5; urutan default `newest`; filter rating divalidasi backend. |
