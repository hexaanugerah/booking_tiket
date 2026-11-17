"use client";

import { useEffect, useState, type FormEvent } from "react";
import { Banknote, Loader2, Search, X } from "lucide-react";
import { toast } from "sonner";
import { bookingService } from "@/services/booking.api.service";
import { refundService } from "@/services/refund.api.service";
import { calculateRefundBreakdown, rupiah } from "@/lib/business";
import type { Booking } from "@/types/api";
import type { Refund } from "@/types/refund";

interface AdminRefundDialogProps {
  open: boolean;
  onClose: () => void;
  onSuccess: (refund: Refund) => void;
}

// Dialog refund langsung oleh admin. Nominal selalu penuh (tanpa potongan).
export function AdminRefundDialog({ open, onClose, onSuccess }: AdminRefundDialogProps) {
  const [query, setQuery] = useState("");
  const [results, setResults] = useState<Booking[]>([]);
  const [searching, setSearching] = useState(false);
  const [selected, setSelected] = useState<Booking | null>(null);
  const [reason, setReason] = useState("");
  const [note, setNote] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!open) return;
    setQuery("");
    setResults([]);
    setSelected(null);
    setReason("");
    setNote("");
    setError(null);
  }, [open]);

  useEffect(() => {
    if (!open) return;

    let alive = true;
    setSearching(true);
    setError(null);

    const timer = setTimeout(() => {
      Promise.all([
        bookingService.list({ page: 1, limit: 10, status: "PAID", search: query }),
        bookingService.list({ page: 1, limit: 10, status: "CONFIRMED", search: query }),
      ])
        .then(([paid, confirmed]) => {
          if (!alive) return;
          const map = new Map<string, Booking>();
          [...paid, ...confirmed].forEach((booking) => map.set(booking.id, booking));
          setResults(Array.from(map.values()));
        })
        .catch((err: any) => {
          if (alive) setError(err?.message || "Gagal mencari booking");
        })
        .finally(() => {
          if (alive) setSearching(false);
        });
    }, 300);

    return () => {
      alive = false;
      clearTimeout(timer);
      setSearching(false);
    };
  }, [open, query]);

  if (!open) return null;

  const breakdown = selected ? calculateRefundBreakdown(selected.total, "ADMIN") : null;

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!selected) return;

    setSubmitting(true);
    setError(null);
    try {
      const refund = await refundService.createByAdmin({
        booking_id: selected.id,
        reason: reason.trim() || undefined,
        admin_note: note.trim() || undefined,
      });
      toast.success("Refund langsung diproses", {
        description: `${refund.booking_code} · ${rupiah(refund.amount)} (tanpa potongan)`,
      });
      onSuccess(refund);
    } catch (err: any) {
      setError(err?.message || "Gagal memproses refund langsung");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div
      className="fixed inset-0 z-50 grid place-items-center bg-black/40 p-4"
      role="dialog"
      aria-modal="true"
      aria-labelledby="admin-refund-dialog-title"
      onClick={onClose}
    >
      <div
        className="max-h-[92vh] w-full max-w-lg overflow-y-auto rounded-md bg-white p-5"
        onClick={(event) => event.stopPropagation()}
      >
        <div className="flex items-start justify-between gap-3 border-b border-[#e5edf3] pb-3">
          <div className="flex items-center gap-2">
            <span className="grid h-9 w-9 place-items-center rounded-full bg-[#e6f7f0] text-[#16876a]">
              <Banknote className="h-4 w-4" aria-hidden="true" />
            </span>
            <div>
              <h3 id="admin-refund-dialog-title" className="text-base font-bold text-[#17324d]">
                Refund Langsung (Admin)
              </h3>
              <p className="text-[11px] text-[#698097]">
                Dana dikembalikan penuh, tanpa potongan 10%.
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

        <form onSubmit={handleSubmit} className="mt-4 space-y-4">
          <div>
            <label className="block">
              <span className="text-xs font-semibold text-[#29445e]">Cari booking</span>
              <div className="relative mt-1">
                <Search
                  className="pointer-events-none absolute left-3 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-[#94A3B8]"
                  aria-hidden="true"
                />
                <input
                  value={query}
                  onChange={(event) => setQuery(event.target.value)}
                  placeholder="Kode booking / nama pengunjung"
                  className="h-9 w-full rounded-md border border-[#d7e3ed] pl-9 pr-3 text-sm text-[#17324d] outline-none placeholder:text-[#94A3B8] focus:border-[#1768ad]"
                />
                {searching && (
                  <Loader2 className="absolute right-3 top-1/2 h-3.5 w-3.5 -translate-y-1/2 animate-spin text-[#698097]" />
                )}
              </div>
            </label>

            <div className="mt-2 max-h-44 overflow-y-auto rounded-md border border-[#e5edf3]">
              {searching && results.length === 0 ? (
                <p className="px-3 py-6 text-center text-xs text-[#698097]">Mencari booking…</p>
              ) : results.length === 0 ? (
                <p className="px-3 py-6 text-center text-xs text-[#94A3B8]">
                  Tidak ada booking PAID/CONFIRMED yang cocok.
                </p>
              ) : (
                <ul className="divide-y divide-[#eef4f8]">
                  {results.map((booking) => {
                    const active = selected?.id === booking.id;
                    return (
                      <li key={booking.id}>
                        <button
                          type="button"
                          onClick={() => setSelected(booking)}
                          className={`w-full px-3 py-2 text-left transition-colors hover:bg-[#f2f7fb] ${
                            active ? "bg-[#e9f3fb]" : "bg-white"
                          }`}
                        >
                          <div className="flex items-center justify-between gap-2">
                            <span className="font-mono text-[11px] font-semibold text-[#17324d]">
                              {booking.booking_code}
                            </span>
                            <span className="text-[11px] font-bold tabular-nums text-[#1768ad]">
                              {rupiah(booking.total)}
                            </span>
                          </div>
                          <p className="mt-0.5 text-[11px] text-[#698097]">
                            {booking.visitor_name} · {booking.visit_date} · {booking.status}
                          </p>
                        </button>
                      </li>
                    );
                  })}
                </ul>
              )}
            </div>
          </div>

          {selected && breakdown && (
            <div className="rounded-md border border-[#d7e3ed] bg-[#f7fbff] p-3 text-xs text-[#29445e]">
              <div className="flex flex-wrap items-center justify-between gap-1">
                <span className="font-semibold">{selected.booking_code}</span>
                <span>{selected.visit_date}</span>
              </div>
              <p className="mt-1 text-[#698097]">
                {selected.total_tickets} tiket · {selected.visitor_name}
              </p>
              <div className="mt-2 flex items-center justify-between gap-2 border-t border-[#d7e3ed] pt-2">
                <span className="font-semibold text-[#17324d]">Nominal refund (penuh)</span>
                <span className="tabular-nums text-sm font-bold text-[#16876a]">
                  {rupiah(breakdown.netAmount)}
                </span>
              </div>
              <p className="mt-1 text-[10px] text-[#16876a]">Tanpa potongan 10%</p>
            </div>
          )}

          <label className="block">
            <span className="text-xs font-semibold text-[#29445e]">
              Alasan / keterangan (opsional)
            </span>
            <textarea
              value={reason}
              onChange={(event) => setReason(event.target.value)}
              maxLength={500}
              rows={2}
              placeholder="Contoh: Pembatalan oleh pengelola."
              className="mt-1 w-full rounded-md border border-[#d7e3ed] px-3 py-2 text-sm text-[#17324d] outline-none placeholder:text-[#94A3B8] focus:border-[#1768ad]"
            />
          </label>

          <label className="block">
            <span className="text-xs font-semibold text-[#29445e]">
              Catatan admin (opsional)
            </span>
            <textarea
              value={note}
              onChange={(event) => setNote(event.target.value)}
              maxLength={500}
              rows={2}
              placeholder="Contoh: Dana dikembalikan melalui transfer manual."
              className="mt-1 w-full rounded-md border border-[#d7e3ed] px-3 py-2 text-sm text-[#17324d] outline-none placeholder:text-[#94A3B8] focus:border-[#1768ad]"
            />
          </label>

          {error && (
            <p className="rounded-md bg-[#fdeaea] px-3 py-2 text-[11px] text-[#c53030]">{error}</p>
          )}

          <p className="rounded-md bg-[#fdf3e3] px-3 py-2 text-[11px] leading-relaxed text-[#8a6420]">
            Refund langsung akan menandai pembayaran sebagai refunded, membatalkan booking,
            dan menonaktifkan e-ticket. Tindakan ini tidak dapat dibatalkan.
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
              disabled={!selected || submitting}
              className="inline-flex h-9 items-center gap-1.5 rounded-md bg-[#16876a] px-4 text-sm font-semibold text-white hover:bg-[#126e57] disabled:opacity-60"
            >
              {submitting && <Loader2 className="h-3.5 w-3.5 animate-spin" />}
              Proses Refund
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
