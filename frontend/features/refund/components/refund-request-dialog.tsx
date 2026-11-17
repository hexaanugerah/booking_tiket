"use client";

import { useEffect, useState, type FormEvent } from "react";
import { Banknote, Landmark, Loader2, Smartphone, X } from "lucide-react";
import { toast } from "sonner";
import { calculateRefundBreakdown, rupiah } from "@/lib/business";
import { refundService } from "@/services/refund.api.service";
import {
  REFUND_BANK_OPTIONS,
  REFUND_EWALLET_OPTIONS,
  refundAccountNameLabel,
  refundAccountNumberLabel,
  type RefundDestinationType,
} from "@/types/refund";
import type { Booking } from "@/types/api";

interface RefundRequestDialogProps {
  booking: Booking | null;
  onClose: () => void;
  onSuccess: () => void;
}

const CUSTOM_PROVIDER = "__OTHER__";

const DESTINATION_OPTIONS = [
  {
    type: "BANK" as const,
    label: "Rekening Bank",
    hint: "BCA, BRI, Mandiri, dll.",
    icon: Landmark,
  },
  {
    type: "EWALLET" as const,
    label: "E-Wallet",
    hint: "DANA, OVO, GoPay, dll.",
    icon: Smartphone,
  },
];

// Dialog pengajuan refund dari halaman Riwayat pengunjung.
export function RefundRequestDialog({ booking, onClose, onSuccess }: RefundRequestDialogProps) {
  const [reason, setReason] = useState("");
  const [destinationType, setDestinationType] = useState<RefundDestinationType>("BANK");
  const [provider, setProvider] = useState("");
  const [customProvider, setCustomProvider] = useState("");
  const [accountName, setAccountName] = useState("");
  const [accountNumber, setAccountNumber] = useState("");
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (!booking) return;
    setReason("");
    setDestinationType("BANK");
    setProvider("");
    setCustomProvider("");
    setAccountName("");
    setAccountNumber("");
  }, [booking]);

  if (!booking) return null;

  const breakdown = calculateRefundBreakdown(booking.total, "USER");
  const providerOptions =
    destinationType === "BANK" ? REFUND_BANK_OPTIONS : REFUND_EWALLET_OPTIONS;
  const isCustomProvider = provider === CUSTOM_PROVIDER;
  const providerName = isCustomProvider ? customProvider.trim() : provider;

  function chooseDestinationType(type: RefundDestinationType) {
    setDestinationType(type);
    setProvider("");
    setCustomProvider("");
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!booking) return;

    if (!providerName) {
      toast.error("Tujuan refund belum lengkap", {
        description: "Pilih bank atau e-wallet tujuan terlebih dahulu.",
      });
      return;
    }

    setSubmitting(true);
    try {
      await refundService.create({
        booking_id: booking.id,
        reason,
        destination_type: destinationType,
        bank_name: providerName,
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
          <div className="mt-2 space-y-1 border-t border-[#d7e3ed] pt-2">
            <div className="flex items-center justify-between gap-2">
              <span className="text-[#698097]">Total pembayaran</span>
              <span className="tabular-nums font-semibold text-[#17324d]">
                {rupiah(breakdown.originalAmount)}
              </span>
            </div>
            <div className="flex items-center justify-between gap-2 text-[#c53030]">
              <span>Potongan {breakdown.deductionPercent}%</span>
              <span className="tabular-nums font-semibold">
                - {rupiah(breakdown.deductionAmount)}
              </span>
            </div>
            <div className="flex items-center justify-between gap-2 border-t border-[#d7e3ed] pt-1">
              <span className="font-semibold text-[#17324d]">Nominal diterima</span>
              <span className="tabular-nums text-sm font-bold text-[#16876a]">
                {rupiah(breakdown.netAmount)}
              </span>
            </div>
          </div>
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

          <div>
            <span className="text-xs font-semibold text-[#29445e]">Tujuan refund</span>
            <div
              className="mt-1 grid grid-cols-2 gap-2"
              role="radiogroup"
              aria-label="Tujuan refund"
            >
              {DESTINATION_OPTIONS.map((option) => {
                const active = destinationType === option.type;
                const Icon = option.icon;
                return (
                  <button
                    key={option.type}
                    type="button"
                    role="radio"
                    aria-checked={active}
                    onClick={() => chooseDestinationType(option.type)}
                    className={`flex items-start gap-2 rounded-md border px-3 py-2 text-left transition-colors ${
                      active
                        ? "border-[#1768ad] bg-[#e9f3fb] text-[#17324d]"
                        : "border-[#d7e3ed] bg-white text-[#29445e] hover:bg-[#f2f7fb]"
                    }`}
                  >
                    <Icon
                      className={`mt-0.5 h-4 w-4 ${
                        active ? "text-[#1768ad]" : "text-[#698097]"
                      }`}
                      aria-hidden="true"
                    />
                    <span>
                      <span className="block text-xs font-semibold">{option.label}</span>
                      <span className="block text-[10px] text-[#698097]">{option.hint}</span>
                    </span>
                  </button>
                );
              })}
            </div>
          </div>

          <div
            className={isCustomProvider ? "grid gap-3 sm:grid-cols-2" : undefined}
          >
            <label className="block">
              <span className="text-xs font-semibold text-[#29445e]">
                {destinationType === "BANK" ? "Bank tujuan" : "E-Wallet tujuan"}
              </span>
              <select
                value={provider}
                onChange={(event) => setProvider(event.target.value)}
                required
                className="mt-1 h-9 w-full rounded-md border border-[#d7e3ed] bg-white px-2 text-sm text-[#17324d] outline-none focus:border-[#1768ad]"
              >
                <option value="">
                  {destinationType === "BANK" ? "Pilih bank" : "Pilih e-wallet"}
                </option>
                {providerOptions.map((option) => (
                  <option key={option} value={option}>
                    {option}
                  </option>
                ))}
                <option value={CUSTOM_PROVIDER}>Lainnya…</option>
              </select>
            </label>

            {isCustomProvider && (
              <label className="block">
                <span className="text-xs font-semibold text-[#29445e]">
                  Nama bank / e-wallet
                </span>
                <input
                  value={customProvider}
                  onChange={(event) => setCustomProvider(event.target.value)}
                  required
                  minLength={2}
                  maxLength={100}
                  placeholder={destinationType === "BANK" ? "Contoh: Bank Mega" : "Contoh: DANA"}
                  className="mt-1 h-9 w-full rounded-md border border-[#d7e3ed] px-3 text-sm text-[#17324d] outline-none placeholder:text-[#94A3B8] focus:border-[#1768ad]"
                />
              </label>
            )}
          </div>

          <div className="grid gap-3 sm:grid-cols-2">
            <label className="block">
              <span className="text-xs font-semibold text-[#29445e]">
                {refundAccountNameLabel(destinationType)}
              </span>
              <input
                value={accountName}
                onChange={(event) => setAccountName(event.target.value)}
                required
                minLength={2}
                maxLength={150}
                placeholder={
                  destinationType === "BANK"
                    ? "Sesuai buku tabungan"
                    : "Sesuai nama akun e-wallet"
                }
                className="mt-1 h-9 w-full rounded-md border border-[#d7e3ed] px-3 text-sm text-[#17324d] outline-none placeholder:text-[#94A3B8] focus:border-[#1768ad]"
              />
            </label>
            <label className="block">
              <span className="text-xs font-semibold text-[#29445e]">
                {refundAccountNumberLabel(destinationType)}
              </span>
              <input
                value={accountNumber}
                onChange={(event) => setAccountNumber(event.target.value)}
                required
                minLength={destinationType === "EWALLET" ? 8 : 4}
                maxLength={50}
                inputMode={destinationType === "EWALLET" ? "tel" : "numeric"}
                placeholder={destinationType === "EWALLET" ? "081234567890" : "1234567890"}
                className="mt-1 h-9 w-full rounded-md border border-[#d7e3ed] px-3 text-sm text-[#17324d] outline-none placeholder:text-[#94A3B8] focus:border-[#1768ad]"
              />
            </label>
          </div>

          <p className="rounded-md bg-[#fdf3e3] px-3 py-2 text-[11px] leading-relaxed text-[#8a6420]">
            Pastikan nama dan nomor tujuan sudah benar. Refund hanya dapat diajukan paling
            lambat H-1 sebelum tanggal kunjungan. Pengajuan refund dari pengunjung dikenakan
            potongan {breakdown.deductionPercent}% dari total pembayaran. Refund yang diproses
            langsung oleh admin tidak dikenakan potongan.
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
