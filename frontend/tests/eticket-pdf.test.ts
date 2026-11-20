import assert from "node:assert/strict";
import { test } from "node:test";
import jsQR from "jsqr";
import { jsPDF } from "jspdf";
import { downloadETicketPDF, toPdfBooking, toPdfTicket } from "../lib/eticket-pdf";
import type { Booking as ApiBooking, ETicket as ApiETicket } from "../types/api";

// Uji regresi QR e-tiket: PDF yang diunduh pengunjung harus memuat payload
// qr_data yang sama dengan QR di halaman e-tiket, sehingga hasil scan petugas
// selalu menunjuk tiket yang sama. Sebelumnya PDF membuat payload JSON sendiri.

const QR_DATA = "KPT-9F8E7D6C5B4A39281706";
const BOOKING_CODE = "KP-20261119-7f3c9a";
const BOOKING_ID = "b1e2d3c4-5f60-4718-8293-a4b5c6d7e8f9";

const apiTicket: ApiETicket = {
  id: "7f3c9a10-2b44-4c8e-9a51-0d6e1f2a3b4c",
  booking_id: BOOKING_ID,
  booking_code: BOOKING_CODE,
  visitor_name: "Budi Santoso",
  email: "budi@example.com",
  visit_date: "2026-11-19",
  ticket_name: "Tiket Dewasa x2",
  quantity: 2,
  total: 100000,
  status: "VALID",
  generated_at: "2026-11-19T08:00:00Z",
  qr_data: QR_DATA,
};

const apiBooking: ApiBooking = {
  id: BOOKING_ID,
  booking_code: BOOKING_CODE,
  visitor_name: apiTicket.visitor_name,
  email: apiTicket.email,
  phone: "08123456789",
  visit_date: apiTicket.visit_date,
  items: [{ ticket_id: "t-1", ticket_name: "Tiket Dewasa", price: 50000, quantity: 2 }],
  subtotal: 100000,
  discount: 0,
  service_fee: 0,
  tax: 0,
  total: 100000,
  total_tickets: 2,
  status: "CONFIRMED",
  created_at: "2026-11-19T07:00:00Z",
  updated_at: "2026-11-19T08:00:00Z",
};

interface EmbeddedImage {
  width: number;
  height: number;
  channels: number;
  pixels: Uint8ClampedArray;
}

// jsPDF menanamkan QR sebagai image stream tanpa kompresi di PDF.
function extractEmbeddedImage(pdf: Buffer): EmbeddedImage {
  const marker = pdf.indexOf("/Subtype /Image");
  assert.notEqual(marker, -1, "PDF harus memuat gambar QR");

  const objectStart = pdf.lastIndexOf(" obj", marker);
  const dictStart = pdf.indexOf("<<", objectStart);
  const streamKeyword = pdf.indexOf("stream", dictStart);
  const dict = pdf.subarray(dictStart, streamKeyword).toString("latin1");

  const width = Number(/\/Width (\d+)/.exec(dict)?.[1]);
  const height = Number(/\/Height (\d+)/.exec(dict)?.[1]);
  const colorSpace = /\/ColorSpace \/(\w+)/.exec(dict)?.[1];
  assert.ok(Number.isFinite(width) && Number.isFinite(height), "dimensi gambar QR harus ada di PDF");
  assert.equal(colorSpace, "DeviceRGB", "format gambar QR berubah — sesuaikan ekstraksi tes");

  let dataStart = streamKeyword + "stream".length;
  if (pdf[dataStart] === 0x0d) dataStart += 1;
  if (pdf[dataStart] === 0x0a) dataStart += 1;

  const channels = 3;
  const byteLength = width * height * channels;
  assert.ok(dataStart + byteLength <= pdf.length, "stream gambar QR terpotong");

  const raw = pdf.subarray(dataStart, dataStart + byteLength);
  const source = new DataView(raw.buffer, raw.byteOffset, raw.byteLength);
  const pixels = new Uint8ClampedArray(width * height * 4);
  const target = new DataView(pixels.buffer);

  for (let i = 0; i < width * height; i++) {
    const src = i * channels;
    const dst = i * 4;
    target.setUint8(dst, source.getUint8(src));
    target.setUint8(dst + 1, source.getUint8(src + 1));
    target.setUint8(dst + 2, source.getUint8(src + 2));
    target.setUint8(dst + 3, 255);
  }

  return { width, height, channels, pixels };
}

test("QR di PDF e-tiket berisi qr_data yang sama dengan halaman e-tiket", async () => {
  let captured: jsPDF | null = null;
  // jsPDF menyalin method dari API ke instance saat konstruksi, jadi save
  // di-stub sebelum lib membuat dokumen agar tidak menulis file ke disk.
  const api = jsPDF.API as typeof jsPDF.API & { save?: (this: jsPDF) => jsPDF };
  const originalSave = api.save;
  api.save = function (this: jsPDF): jsPDF {
    captured = this;
    return this;
  };

  try {
    await downloadETicketPDF(toPdfTicket(apiTicket), toPdfBooking(apiBooking));
  } finally {
    api.save = originalSave;
  }

  assert.ok(captured, "downloadETicketPDF harus menghasilkan dokumen PDF");
  const pdf = Buffer.from((captured as jsPDF).output("arraybuffer"));
  const image = extractEmbeddedImage(pdf);

  // jsQR adalah decoder yang sama dengan kamera di halaman petugas.
  const decoded = jsQR(image.pixels, image.width, image.height, { inversionAttempts: "dontInvert" });
  assert.ok(decoded, "QR di dalam PDF harus bisa dipindai");
  assert.equal(decoded.data, QR_DATA, "payload QR PDF harus sama dengan qr_data");
  assert.ok(!decoded.data.startsWith("{"), "payload QR tidak boleh berupa JSON lama");

  // Kode booking yang ditampilkan adalah kode ramah pengunjung, bukan UUID.
  const pdfText = pdf.toString("latin1");
  assert.ok(pdfText.includes(BOOKING_CODE), "PDF harus menampilkan kode booking");
  assert.ok(!pdfText.includes(BOOKING_ID), "PDF tidak boleh menampilkan UUID booking");

  // Kode tiket yang ditampilkan adalah payload QR (KPT-...), bukan UUID tiket.
  assert.ok(pdfText.includes(QR_DATA), "PDF harus menampilkan kode tiket KPT-...");
  assert.ok(!pdfText.includes(apiTicket.id), "PDF tidak boleh menampilkan UUID tiket");
});
