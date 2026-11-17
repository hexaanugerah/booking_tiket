"use client";

import { useEffect, useState, type FormEvent } from "react";
import { Banknote, Loader2, X } from "lucide-react";
import { toast } from "sonner";
import { rupiah } from "@/lib/business";
import { refundService } from "@/services/refund.api.service";
import type { Booking } from "@/types/api";

interface RefundRequestDialogProps {
  booking: Booking | null;
  onClose: () => void;
  onSuccess: () => void;
}

// Dialog pengajuan refund dari halaman Riwayat pengunjung.
export function RefundRequestDialog({ booking, onClose, onSuccess }: RefundRequestDialogProps) {
  const [reason, setReason] = useState("");
  const [bankName, setBankName] = useState("");
  const [accountName, setAccountName] = useState("");
  const [accountNumber, setAccountNumber] = useState("");
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (!booking) return;
    setReason("");
    setBankName("");
    setAccountName("");
    setAccountNumber("");
  }, [booking]);

  if (!booking) return null;

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!booking) return;

    setSubmitting(true);
    try {
      await refundService.create({
        booking_id: booking.id,
        reason,
        bank_name: bankName,
        account_name: accountName,
        account_number: accountNumber,
      });
      toast.success("Pengajuan refund dikirim", {
        description: "Admin akan meninjau pengajuan Anda terlebih dahulu.",
      });
      onSuccess();
    } catch (error: any) {
      toast.error("Gagal mengirim refund", { description: error.message });
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div
      className="fixed inset-0 z-50 grid place-items-center bg-black/40 p-4"
      role="dialog"
      aria-modal="true"
      aria-labelledby="refund-dialog-title"
      onClick={onClose}
    >
      <div
        className="max-h-[92vh] w-full max-w-lg overflow-y-auto rounded-md bg-white p-5"
        onClick={(event) => event.stopPropagation()}
      >
        <div className="flex items-start justify-between gap-3 border-b border-[#e5edf3] pb-3">
          <div className="flex items-center gap-2">
            <span className="grid h-9 w-9 place-items-center rounded-full bg-[#fdf3e3] text-[#b7791f]">
              <Banknote className="h-4 w-4" aria-hidden="true" />
            </span>
            <div>
              <h3 id="refund-dialog-title" className="text-base font-bold text-[#17324d]">
                Ajukan Pengembalian Dana
              </h3>
              <p className="text-[11px] text-[#698097]">
                Pengajuan akan ditinjau oleh admin.
              </p>
            </div>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="text-[#698097] hover:text-[#17324d]"
            aria-label="Tutup"
          >
            <X className="h-4 w-4" />
          </button>
        </div>

        <div className="mt-3 rounded-md border border-[#d7e3ed] bg-[#f7fbff] p-3 text-xs text-[#29445e]">
          <div className="flex flex-wrap items-center justify-between gap-1">
            <span className="font-semibold">{booking.booking_code}</span>
            <span>{booking.visit_date}</span>
          </div>
          <p className="mt-1 text-[#698097]">
            {booking.total_tickets} tiket · {booking.visitor_name}
          </p>
          <p className="mt-2 text-sm font-bold text-[#17324d]">
            Nominal refund: {rupiah(booking.total)}
          </p>
        </div>

        <form onSubmit={handleSubmit} className="mt-4 space-y-3">
          <label className="block">
            <span className="text-xs font-semibold text-[#29445e]">Alasan refund</span>
            <textarea
              value={reason}
              onChange={(event) => setReason(event.target.value)}
              required
              minLength={10}
              maxLength={500}
              rows={3}
              placeholder="Contoh: Rencana kunjungan berubah karena sakit."
              className="mt-1 w-full rounded-md border border-[#d7e3ed] px-3 py-2 text-sm text-[#17324d] outline-none placeholder:text-[#94A3B8] focus:border-[#1768ad]"
            />
          </label>

          <div className="grid gap-3 sm:grid-cols-2">
            <label className="block">
              <span className="text-xs font-semibold text-[#29445e]">Nama bank</span>
              <input
                value={bankName}
                onChange={(event) => setBankName(event.target.value)}
                required
                minLength={2}
                maxLength={100}
                placeholder="BCA / BRI / Mandiri"
                className="mt-1 h-9 w-full rounded-md border border-[#d7e3ed] px-3 text-sm text-[#17324d] outline-none placeholder:text-[#94A3B8] focus:border-[#1768ad]"
              />
            </label>
            <label className="block">
              <span className="text-xs font-semibold text-[#29445e]">Nama pemilik rekening</span>
              <input
                value={accountName}
                onChange={(event) => setAccountName(event.target.value)}
                required
                minLength={2}
                maxLength={150}
                placeholder="Sesuai buku tabungan"
                className="mt-1 h-9 w-full rounded-md border border-[#d7e3ed] px-3 text-sm text-[#17324d] outline-none placeholder:text-[#94A3B8] focus:border-[#1768ad]"
              />
            </label>
          </div>

          <label className="block">
            <span className="text-xs font-semibold text-[#29445e]">Nomor rekening</span>
            <input
              value={accountNumber}
              onChange={(event) => setAccountNumber(event.target.value)}
              required
              minLength={4}
              maxLength={50}
              inputMode="numeric"
              placeholder="1234567890"
              className="mt-1 h-9 w-full rounded-md border border-[#d7e3ed] px-3 text-sm text-[#17324d] outline-none placeholder:text-[#94A3B8] focus:border-[#1768ad]"
            />
          </label>

          <p className="rounded-md bg-[#fdf3e3] px-3 py-2 text-[11px] leading-relaxed text-[#8a6420]">
            Refund hanya dapat diajukan paling lambat H-1 sebelum tanggal kunjungan.
            Nominal yang dikembalikan mengikuti total pembayaran booking.
          </p>

          <div className="flex justify-end gap-2 pt-1">
            <button
              type="button"
              onClick={onClose}
              className="h-9 rounded-md border border-[#d7e3ed] px-4 text-sm font-semibold text-[#29445e] hover:bg-[#f2f7fb]"
            >
              Batal
            </button>
            <button
              type="submit"
              disabled={submitting}
              className="inline-flex h-9 items-center gap-1.5 rounded-md bg-[#1768ad] px-4 text-sm font-semibold text-white hover:bg-[#14548f] disabled:opacity-60"
            >
              {submitting && <Loader2 className="h-3.5 w-3.5 animate-spin" />}
              Kirim Pengajuan
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
