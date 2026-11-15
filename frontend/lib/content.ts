// Konten statis — teks/marketing statis untuk halaman publik. Ini BUKAN data mock database; data operasional selalu diambil dari API.

export interface AboutSection {
  key: string;
  title: string;
  body: string;
}

export interface HomeTicketOption {
  key: string;
  title: string;
  desc: string;
  price: string;
  quantity: number;
}

// Kartu ringkas "Pilih Tiketmu" di beranda — mengacu pada Price List resmi
// Wana Wisata Kawah Putih (lihat backend/internal/seed/seed.go).
export const homeTicketOptions: HomeTicketOption[] = [
  { key: "htm", title: "HTM Wisatawan Nusantara", desc: "Tiket masuk kawasan Kawah Putih", price: "Rp31.000", quantity: 0 },
  { key: "kendaraan", title: "Kendaraan Roda 4", desc: "Biaya masuk kendaraan roda 4", price: "Rp18.000", quantity: 0 },
  { key: "wahana", title: "Ontang Anting", desc: "Wahana & aktivitas pilihan", price: "Rp32.000", quantity: 0 },
];

export const aboutHistory =
  "Kawah Putih merupakan danau kawah vulkanik yang berada di kawasan Gunung Patuha, Ciwidey, Jawa Barat. Danau ini terbentuk dari aktivitas vulkanik dan memiliki air berwarna putih kehijauan yang dapat berubah mengikuti kondisi alam. Keunikan warna dan suasana berkabutnya menjadikan Kawah Putih salah satu destinasi wisata alam yang menarik di Jawa Barat.";

export const aboutSections: AboutSection[] = [
  { key: "fasilitas", title: "Fasilitas", body: "Area parkir luas, pusat informasi wisata, musala, toilet bersih, kafetaria, serta jalur pejalan kaki yang aman menuju bibir kawah." },
  { key: "informasi", title: "Informasi Destinasi", body: "Informasi destinasi diperbarui secara berkala — termasuk jam operasional, harga tiket, dan kondisi cuaca terkini." },
  { key: "aturan", title: "Aturan Kunjungan", body: "Jaga kebersihan kawasan, ikuti jalur yang tersedia, dilarang berenang, dan selalu patuhi arahan petugas di lapangan." },
];
