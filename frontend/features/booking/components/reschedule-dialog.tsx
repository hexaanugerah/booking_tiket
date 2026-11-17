"use client";

import { useEffect, useState, type FormEvent } from "react";
import { CalendarDays, Loader2, X } from "lucide-react";
import { toast } from "sonner";
import { addDaysISO, todayISO } from "@/lib/business";
import { rescheduleService } from "@/services/reschedule.api.service";
import {
  ACTIVE_RESCHEDULE_STATUSES,
  MAX_RESCHEDULES_PER_BOOKING,
  type Reschedule,
} from "@/types/reschedule";
import type { Booking } from "@/types/api";

interface RescheduleDialogProps {
  booking: Booking | null;
  onClose: () => void;
  onSuccess: () => void;
}

// Dialog pengajuan ubah tanggal kunjungan dari halaman Riwayat pengunjung.
// Tanggal booking baru berubah setelah pengajuan disetujui admin.
export function RescheduleDialog({ booking, onClose, onSuccess }: RescheduleDialogProps) {
  const [newDate, setNewDate] = useState("");
  const [reason, setReason] = useState("");
  const [history, setHistory] = useState<Reschedule[] | null>(null);
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (!booking) return;
    setNewDate("");
    setReason("");
    setHistory(null);

    let alive = true;
    rescheduleService
      .list(booking.id)
      .then((rows) => {
        if (alive) setHistory(rows);
      })
      .catch(() => {
        // Riwayat hanya untuk info jatah; backend tetap memvalidasi.
        if (alive) setHistory([]);
      });
    return () => {
      alive = false;
    };
  }, [booking]);

  if (!booking) return null;

  const minimumDate = addDaysISO(todayISO(), 1);
  const activeReschedules = (history ?? []).filter((item) =>
    ACTIVE_RESCHEDULE_STATUSES.includes(item.status)
  );
  const pendingRequest = history?.find((item) => item.status === "PENDING") ?? null;
  const usedCount = activeReschedules.length;
  const limitReached =
    history !== null && usedCount >= MAX_RESCHEDULES_PER_BOOKING;

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!booking) return;

    if (pendingRequest) {
      toast.error("Masih menunggu persetujuan", {
        description: "Pengajuan sebelumnya belum diproses admin.",
      });
      return;
    }
    if (limitReached) {
      toast.error("Batas jadwal ulang tercapai", {
        description: `Maksimal ${MAX_RESCHEDULES_PER_BOOKING} kali per booking.`,
      });
      return;
    }
    if (newDate === booking.visit_date) {
      toast.error("Pilih tanggal yang berbeda", {
        description: "Tanggal baru tidak boleh sama dengan tanggal kunjungan saat ini.",
      });
      return;
    }

    setSubmitting(true);
    try {
      await rescheduleService.create(booking.id, {
        new_visit_date: newDate,
        reason,
      });
      toast.success("Pengajuan jadwal ulang dikirim", {
        description: `Tanggal ${newDate} menunggu persetujuan admin.`,
      });
      onSuccess();
    } catch (error: any) {
      toast.error("Gagal mengajukan jadwal ulang", { description: error.message });
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div
      className="fixed inset-0 z-50 grid place-items-center bg-black/40 p-4"
      role="dialog"
      aria-modal="true"
      aria-labelledby="reschedule-dialog-title"
      onClick={onClose}
    >
      <div
        className="w-full max-w-md rounded-md bg-white p-5"
        onClick={(event) => event.stopPropagation()}
      >
        <div className="flex items-start justify-between gap-3 border-b border-[#e5edf3] pb-3">
          <div className="flex items-center gap-2">
            <span className="grid h-9 w-9 place-items-center rounded-full bg-[#e9f3fb] text-[#1768ad]">
              <CalendarDays className="h-4 w-4" aria-hidden="true" />
            </span>
            <div>
              <h3 id="reschedule-dialog-title" className="text-base font-bold text-[#17324d]">
                Ajukan Jadwal Ulang
              </h3>
              <p className="text-[11px] text-[#698097]">{booking.booking_code}</p>
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
          <p>
            Tanggal kunjungan saat ini:{" "}
            <span className="font-bold text-[#17324d]">{booking.visit_date}</span>
          </p>
          <p className="mt-1 text-[#698097]">
            {booking.total_tickets} tiket · jadwal ulang terpakai{" "}
            <span className="font-semibold text-[#29445e]">
              {history === null
                ? "…"
                : `${usedCount}/${MAX_RESCHEDULES_PER_BOOKING}`}
            </span>
            .
          </p>
        </div>

        {pendingRequest && (
          <p className="mt-2 rounded-md bg-[#fdf3e3] px-3 py-2 text-[11px] leading-relaxed text-[#8a6420]">
            Pengajuan ke tanggal{" "}
            <span className="font-semibold">{pendingRequest.new_visit_date}</span> masih
            menunggu persetujuan admin. Ajukan lagi setelah pengajuan ini diproses.
          </p>
        )}

        {limitReached && (
          <p className="mt-2 rounded-md bg-[#fdeaea] px-3 py-2 text-[11px] text-[#c53030]">
            Batas maksimal {MAX_RESCHEDULES_PER_BOOKING} kali jadwal ulang sudah tercapai
            untuk booking ini.
          </p>
        )}

        <form onSubmit={handleSubmit} className="mt-4 space-y-3">
          <label className="block">
            <span className="text-xs font-semibold text-[#29445e]">Tanggal kunjungan baru</span>
            <input
              type="date"
              value={newDate}
              onChange={(event) => setNewDate(event.target.value)}
              min={minimumDate}
              required
              className="mt-1 h-9 w-full rounded-md border border-[#d7e3ed] px-3 text-sm text-[#17324d] outline-none focus:border-[#1768ad]"
            />
          </label>

          <label className="block">
            <span className="text-xs font-semibold text-[#29445e]">Alasan jadwal ulang</span>
            <textarea
              value={reason}
              onChange={(event) => setReason(event.target.value)}
              required
              minLength={10}
              maxLength={500}
              rows={3}
              placeholder="Contoh: Ada acara keluarga pada tanggal sebelumnya."
              className="mt-1 w-full rounded-md border border-[#d7e3ed] px-3 py-2 text-sm text-[#17324d] outline-none placeholder:text-[#94A3B8] focus:border-[#1768ad]"
            />
          </label>

          <p className="rounded-md bg-[#fdf3e3] px-3 py-2 text-[11px] leading-relaxed text-[#8a6420]">
            Pengajuan akan ditinjau admin. Tanggal kunjungan baru berlaku setelah
            disetujui dan tergantung ketersediaan kuota tiket pada tanggal tersebut.
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
              disabled={submitting || limitReached || pendingRequest !== null}
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
