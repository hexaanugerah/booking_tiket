"use client";

import { jsPDF } from "jspdf";
import QRCode from "qrcode";
import { rupiah } from "@/lib/business";
import type { Booking, ETicket } from "@/types/domain";
import type { Booking as ApiBooking, ETicket as ApiETicket } from "@/types/api";

// E-ticket PDF — generate PDF e-tiket di sisi klien (jsPDF + QR PNG). Dipanggil dari tombol Download PDF.

// Pemetaan tunggal dari respons API (snake_case) ke tipe domain PDF.
// Semua jalur unduh wajib memakai fungsi ini supaya payload QR selalu sama
// dengan qr_data yang tampil di halaman e-tiket.
export function toPdfTicket(t: ApiETicket): ETicket {
  return {
    ...t,
    bookingId: t.booking_id,
    visitorName: t.visitor_name,
    visitDate: t.visit_date,
    ticketName: t.ticket_name,
    qrData: t.qr_data,
    generatedAt: t.generated_at,
    checkedInAt: t.checked_in_at,
    checkedInGate: t.checked_in_gate,
  };
}

// Bungkus nilai ke lebar kolom. Kata yang lebih lebar dari kolom (mis. email
// panjang) dipotong di tanda baca terdekat supaya rapi, bukan di tengah kata.
function wrapText(doc: jsPDF, text: string, width: number): string[] {
  type Unit = { text: string; gap: boolean };
  const units: Unit[] = [];
  for (const word of text.split(/\s+/)) {
    if (!word) continue;
    let first = true;
    // Simpan tanda baca di akhir potongan: "budi." "santoso." "com"
    for (const part of word.split(/(?<=[.@\-_/+])/)) {
      if (!part) continue;
      units.push({ text: part, gap: first });
      first = false;
    }
  }

  const lines: string[] = [];
  let current = "";
  for (const unit of units) {
    const candidate = current ? current + (unit.gap ? " " : "") + unit.text : unit.text;
    if (doc.getTextWidth(candidate) <= width) {
      current = candidate;
      continue;
    }
    if (current) {
      lines.push(current);
      current = "";
    }
    // Potongan tanpa tanda baca masih terlalu lebar: potong per karakter.
    let chunk = "";
    for (const ch of unit.text) {
      if (doc.getTextWidth(chunk + ch) <= width) {
        chunk += ch;
      } else {
        lines.push(chunk);
        chunk = ch;
      }
    }
    current = chunk;
  }
  if (current) lines.push(current);
  return lines.length ? lines : [""];
}

export function toPdfBooking(b: ApiBooking): Booking {
  return {
    ...b,
    bookingCode: b.booking_code,
    userId: b.user_id ?? null,
    visitorName: b.visitor_name,
    visitDate: b.visit_date,
    items: b.items.map((it) => ({
      ticketId: it.ticket_id,
      ticketName: it.ticket_name,
      price: it.price,
      quantity: it.quantity,
    })),
    serviceFee: b.service_fee,
    totalTickets: b.total_tickets,
    createdAt: b.created_at,
    ticketId: b.ticket_id ?? null,
  };
}

export async function downloadETicketPDF(ticket: ETicket, booking: Booking) {
  const doc = new jsPDF({ unit: "pt", format: "a4", orientation: "portrait" });
  const W = doc.internal.pageSize.getWidth();

  // Header brand
  doc.setFillColor(23, 104, 173); // #1768ad
  doc.rect(0, 0, W, 90, "F");
  doc.setTextColor(255, 255, 255);
  doc.setFont("helvetica", "bold");
  doc.setFontSize(20);
  doc.text("KAWAH PUTIH RANCABALI", 40, 42);
  doc.setFont("helvetica", "normal");
  doc.setFontSize(11);
  doc.text("E-Tiket Kunjungan — Jl. Raya Soreang-Ciwidey KM.25, Bandung", 40, 62);

  // Judul & status
  let y = 130;
  doc.setTextColor(23, 50, 77); // #17324d
  doc.setFont("helvetica", "bold");
  doc.setFontSize(16);
  doc.text("E-TIKET", 40, y);

  doc.setFontSize(10);
  const statusColor: Record<string, [number, number, number]> = {
    VALID: [22, 135, 106],
    USED: [23, 104, 173],
    EXPIRED: [181, 76, 26],
    CANCELLED: [197, 48, 48],
  };
  const [r, g, b] = statusColor[ticket.status] ?? [23, 50, 77];
  doc.setTextColor(r, g, b);
  doc.text(`STATUS: ${ticket.status}`, W - 40, y, { align: "right" });

  // QR Code — payload harus identik dengan QR yang tampil di halaman e-tiket
  // (qr_data dari backend) supaya hasil scan sama persis.
  const qrPayload = ticket.qrData || ticket.id;
  const qrDataUrl = await QRCode.toDataURL(qrPayload, { width: 360, margin: 1 });
  const qrSize = 150;
  const qrX = W - 40 - qrSize;
  doc.addImage(qrDataUrl, "PNG", qrX, y + 16, qrSize, qrSize);

  // Info booking
  doc.setTextColor(23, 50, 77);
  const rows: [string, string][] = [
    ["Kode Booking", booking.bookingCode || booking.id],
    // Kode tiket yang dibaca petugas adalah payload QR (KPT-...), bukan UUID internal.
    ["Kode Tiket", ticket.qrData || ticket.id],
    ["Nama Pemesan", booking.visitorName],
    ["Email", booking.email],
    ["Tanggal Kunjungan", booking.visitDate],
    ["Jenis Tiket", ticket.ticketName],
    ["Jumlah Tiket", `${ticket.quantity} tiket`],
    ["Total Pembayaran", rupiah(booking.total)],
    ["Metode Pembayaran", "QRIS"],
    ["Dibuat", new Date(ticket.generatedAt).toLocaleString("id-ID")],
  ];

  y += 20;
  doc.setFontSize(11);
  // Nilai dibungkus dalam kolom kiri (sampai tepi QR dikurangi jarak aman),
  // supaya nama/email/jenis tiket yang panjang tidak menabrak QR.
  const labelX = 40;
  const valueX = 160;
  const valueWidth = qrX - 14 - valueX;
  const lineHeight = 14;
  const rowGap = 8; // 14 + 8 = 22, sama dengan jarak baris sebelumnya
  let ry = y;
  rows.forEach(([label, value]) => {
    doc.setFont("helvetica", "bold");
    const lines = wrapText(doc, String(value), valueWidth);
    doc.setFont("helvetica", "normal");
    doc.setTextColor(105, 128, 151);
    doc.text(label, labelX, ry);
    doc.setFont("helvetica", "bold");
    doc.setTextColor(23, 50, 77);
    doc.text(lines, valueX, ry);
    ry += lineHeight * Math.max(1, lines.length) + rowGap;
  });

  // Catatan
  const noteY = ry + 22;
  doc.setDrawColor(215, 227, 237);
  doc.roundedRect(40, noteY, W - 80 - qrSize - 20, 70, 6, 6, "S");
  doc.setFont("helvetica", "normal");
  doc.setFontSize(9);
  doc.setTextColor(105, 128, 151);
  doc.text("Tunjukkan e-tiket ini (cetak atau dari layar) saat tiba di gerbang.", 52, noteY + 20);
  doc.text("Tiket hanya berlaku pada tanggal kunjungan yang tertera.", 52, noteY + 36);
  doc.text("Tiket yang sudah digunakan tidak dapat dipakai kembali.", 52, noteY + 52);

  // Footer
  doc.setFontSize(8);
  doc.setTextColor(148, 163, 184);
  doc.text("Dokumen ini dihasilkan otomatis oleh sistem Kawah Putih Rancabali.", 40, 800);

  doc.save(`e-tiket-${booking.bookingCode || booking.id}.pdf`);
}
