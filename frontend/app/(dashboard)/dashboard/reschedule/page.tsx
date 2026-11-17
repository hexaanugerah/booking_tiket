"use client";

import { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { CalendarClock, CalendarDays, Loader2 } from "lucide-react";
import { PageHeader, SectionCard, StatCard, StatGrid } from "@/components/panel/panel-ui";
import { RescheduleDialog } from "@/features/booking/components/reschedule-dialog";
import { bookingService } from "@/services/booking.api.service";
import { rescheduleService } from "@/services/reschedule.api.service";
import { todayISO } from "@/lib/business";
import {
  ACTIVE_RESCHEDULE_STATUSES,
  MAX_RESCHEDULES_PER_BOOKING,
  RESCHEDULE_STATUS_LABEL,
  type Reschedule,
  type RescheduleStatus,
} from "@/types/reschedule";
import type { Booking } from "@/types/api";

const STATUS_CLASS: Record<RescheduleStatus, string> = {
  PENDING: "bg-[#fdf3e3] text-[#b7791f]",
  APPROVED: "bg-[#e6f7f0] text-[#16876a]",
  REJECTED: "bg-[#fdeaea] text-[#c53030]",
};

// Jadwal Ulang — pengunjung dapat mengajukan pemindahan tanggal kunjungan.
// Maksimal 2 kali per booking, menunggu persetujuan admin, dan tergantung kuota.
export default function MyReschedulePage() {
  const [bookings, setBookings] = useState<Booking[] | null>(null);
  const [reschedules, setReschedules] = useState<Reschedule[]>([]);
  const [selected, setSelected] = useState<Booking | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [refreshKey, setRefreshKey] = useState(0);

  useEffect(() => {
    let alive = true;
    setLoading(true);
    setError(null);

    Promise.all([bookingService.listMine({ limit: 100 }), rescheduleService.listMine()])
      .then(([bookingRows, rescheduleRows]) => {
        if (!alive) return;
        setBookings(bookingRows);
        setReschedules(rescheduleRows);
      })
      .catch((err: any) => {
        if (!alive) return;
        setError(err?.message || "Gagal memuat data jadwal ulang");
      })
      .finally(() => {
        if (alive) setLoading(false);
      });

    return () => {
      alive = false;
    };
  }, [refreshKey]);

  const today = todayISO();

  const countByBooking = useMemo(() => {
    const map = new Map<string, number>();
    for (const item of reschedules) {
      if (!ACTIVE_RESCHEDULE_STATUSES.includes(item.status)) continue;
      map.set(item.booking_id, (map.get(item.booking_id) ?? 0) + 1);
    }
    return map;
  }, [reschedules]);

  const pendingByBooking = useMemo(() => {
    const map = new Map<string, Reschedule>();
    for (const item of reschedules) {
      if (item.status === "PENDING") map.set(item.booking_id, item);
    }
    return map;
  }, [reschedules]);

  const eligibleBookings = useMemo(
    () =>
      (bookings ?? [])
        .filter(
          (booking) =>
            (booking.status === "PAID" || booking.status === "CONFIRMED") &&
            booking.visit_date > today
        )
        .sort((a, b) => a.visit_date.localeCompare(b.visit_date)),
    [bookings, today]
  );

  const rescheduledBookingCount = useMemo(
    () => new Set(reschedules.map((item) => item.booking_id)).size,
    [reschedules]
  );

  return (
    <div>
      <PageHeader
        title="Jadwal Ulang"
        subtitle="Pindahkan tanggal kunjungan tanpa membuat booking baru"
      />

      <StatGrid cols={3}>
        <StatCard label="Booking Bisa Dijadwalkan" value={String(eligibleBookings.length)} />
        <StatCard label="Total Jadwal Ulang" value={String(reschedules.length)} />
        <StatCard label="Booking Sudah Dijadwalkan" value={String(rescheduledBookingCount)} accent />
      </StatGrid>

      {error && (
        <p className="mt-3 rounded-md bg-[#fdeaea] px-3 py-2 text-sm text-[#c53030]">{error}</p>
      )}

      <SectionCard
        title="Booking yang Bisa Dijadwalkan Ulang"
        subtitle={`Maksimal ${MAX_RESCHEDULES_PER_BOOKING} kali per booking dan tergantung ketersediaan kuota tiket.`}
        className="mt-3"
      >
        {loading && bookings === null ? (
          <div className="flex items-center gap-2 py-8 text-sm text-[#698097]">
            <Loader2 className="h-4 w-4 animate-spin" /> Memuat booking…
          </div>
        ) : eligibleBookings.length === 0 ? (
          <div className="py-8 text-center">
            <CalendarClock className="mx-auto h-10 w-10 text-[#94A3B8]" />
            <p className="mt-2 text-sm text-[#698097]">
              Tidak ada booking yang bisa dijadwalkan ulang.
            </p>
            <p className="mx-auto mt-1 max-w-md text-xs leading-relaxed text-[#94A3B8]">
              Jadwal ulang hanya tersedia untuk booking berstatus PAID/CONFIRMED dengan
              tanggal kunjungan minimal besok.
            </p>
            <Link
              href="/dashboard/booking"
              className="mt-3 inline-flex h-8 items-center gap-1.5 rounded-md border border-[#d7e3ed] px-3 text-xs font-semibold text-[#29445e] hover:bg-[#f2f7fb]"
            >
              <CalendarDays className="h-3.5 w-3.5" /> Lihat Booking Saya
            </Link>
          </div>
        ) : (
          <div className="grid gap-3 md:grid-cols-2">
            {eligibleBookings.map((booking) => {
              const used = countByBooking.get(booking.id) ?? 0;
              const pending = pendingByBooking.get(booking.id) ?? null;
              const limitReached = used >= MAX_RESCHEDULES_PER_BOOKING;

              return (
                <div
                  key={booking.id}
                  className="rounded-md border border-[#d7e3ed] bg-[#f7fbff] p-3"
                >
                  <div className="flex flex-wrap items-center justify-between gap-1">
                    <span className="font-mono text-xs font-semibold text-[#17324d]">
                      {booking.booking_code}
                    </span>
                    <span className="text-xs text-[#698097]">{booking.visit_date}</span>
                  </div>
                  <p className="mt-1 text-xs text-[#698097]">
                    {booking.total_tickets} tiket · {booking.visitor_name}
                  </p>

                  <div className="mt-2 flex items-center justify-between gap-2 border-t border-[#d7e3ed] pt-2 text-xs">
                    <span className="text-[#698097]">Jadwal ulang terpakai</span>
                    <span
                      className={`font-semibold tabular-nums ${
                        limitReached ? "text-[#c53030]" : "text-[#17324d]"
                      }`}
                    >
                      {used}/{MAX_RESCHEDULES_PER_BOOKING}
                    </span>
                  </div>

                  {pending && (
                    <p className="mt-2 rounded-md bg-[#fdf3e3] px-2 py-1.5 text-[10px] leading-relaxed text-[#8a6420]">
                      Menunggu persetujuan admin untuk tanggal{" "}
                      <span className="font-semibold">{pending.new_visit_date}</span>.
                    </p>
                  )}

                  <button
                    type="button"
                    onClick={() => setSelected(booking)}
                    disabled={limitReached || pending !== null}
                    className="mt-3 inline-flex h-8 w-full items-center justify-center gap-1.5 rounded-md bg-[#1768ad] px-3 text-xs font-semibold text-white hover:bg-[#14548f] disabled:cursor-not-allowed disabled:opacity-50"
                  >
                    <CalendarClock className="h-3.5 w-3.5" />
                    {pending
                      ? "Menunggu Persetujuan"
                      : limitReached
                        ? "Jatah Jadwal Ulang Habis"
                        : "Ajukan Jadwal Ulang"}
                  </button>
                </div>
              );
            })}
          </div>
        )}
      </SectionCard>

      <SectionCard
        title="Riwayat Jadwal Ulang"
        subtitle="Semua perubahan tanggal kunjungan yang sudah tersimpan."
        className="mt-3"
      >
        <div className="panel-scroll rounded-md border border-[#e5edf3]">
          <table className="panel-table">
            <thead>
              <tr>
                <th>Kode Booking</th>
                <th>Tanggal Lama</th>
                <th>Tanggal Baru</th>
                <th>Alasan</th>
                <th>Status</th>
                <th>Diajukan</th>
              </tr>
            </thead>
            <tbody>
              {loading ? (
                <tr>
                  <td colSpan={6} className="py-10 text-center text-sm text-[#698097]">
                    Memuat…
                  </td>
                </tr>
              ) : reschedules.length === 0 ? (
                <tr>
                  <td colSpan={6} className="py-10 text-center text-sm text-[#94A3B8]">
                    Belum ada jadwal ulang.
                  </td>
                </tr>
              ) : (
                reschedules.map((item) => (
                  <tr key={item.id}>
                    <td className="font-mono text-[11px] font-semibold text-[#17324d]">
                      {item.booking_code}
                    </td>
                    <td className="tabular-nums">{item.old_visit_date}</td>
                    <td className="tabular-nums font-semibold text-[#17324d]">
                      {item.new_visit_date}
                    </td>
                    <td className="max-w-[260px]">
                      <p className="line-clamp-2 text-[11px] text-[#29445e]">{item.reason}</p>
                    </td>
                    <td>
                      <span
                        className={`inline-block rounded-full px-2 py-0.5 text-[10px] font-bold ${STATUS_CLASS[item.status]}`}
                      >
                        {RESCHEDULE_STATUS_LABEL[item.status]}
                      </span>
                      {item.admin_note && (
                        <p className="mt-1 max-w-[200px] text-[10px] leading-relaxed text-[#698097]">
                          {item.admin_note}
                        </p>
                      )}
                    </td>
                    <td className="text-[11px] text-[#698097]">
                      {formatDate(item.created_at)}
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </SectionCard>

      <RescheduleDialog
        booking={selected}
        onClose={() => setSelected(null)}
        onSuccess={() => {
          setSelected(null);
          setRefreshKey((key) => key + 1);
        }}
      />
    </div>
  );
}

function formatDate(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return date.toLocaleDateString("id-ID", {
    day: "2-digit",
    month: "short",
    year: "numeric",
  });
}
