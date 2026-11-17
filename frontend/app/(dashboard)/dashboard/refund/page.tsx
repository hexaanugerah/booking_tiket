"use client";

import { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { Banknote, CalendarDays, Loader2 } from "lucide-react";
import { PageHeader, SectionCard, StatCard, StatGrid } from "@/components/panel/panel-ui";
import { RefundRequestDialog } from "@/features/refund/components/refund-request-dialog";
import { bookingService } from "@/services/booking.api.service";
import { refundService } from "@/services/refund.api.service";
import { calculateRefundBreakdown, rupiah, todayISO } from "@/lib/business";
import {
  REFUND_DESTINATION_LABEL,
  REFUND_INITIATOR_LABEL,
  REFUND_STATUS_LABEL,
  type Refund,
  type RefundStatus,
} from "@/types/refund";
import type { Booking } from "@/types/api";

const STATUS_CLASS: Record<RefundStatus, string> = {
  PENDING: "bg-[#fdf3e3] text-[#b7791f]",
  APPROVED: "bg-[#e9f3fb] text-[#1768ad]",
  REJECTED: "bg-[#fdeaea] text-[#c53030]",
  REFUNDED: "bg-[#e6f7f0] text-[#16876a]",
};

export default function MyRefundPage() {
  const [bookings, setBookings] = useState<Booking[] | null>(null);
  const [refunds, setRefunds] = useState<Refund[]>([]);
  const [selected, setSelected] = useState<Booking | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [refreshKey, setRefreshKey] = useState(0);

  useEffect(() => {
    let alive = true;
    setLoading(true);
    setError(null);

    Promise.all([bookingService.listMine({ limit: 100 }), refundService.listMine()])
      .then(([bookingRows, refundRows]) => {
        if (!alive) return;
        setBookings(bookingRows);
        setRefunds(refundRows);
      })
      .catch((err: any) => {
        if (!alive) return;
        setError(err?.message || "Gagal memuat data refund");
      })
      .finally(() => {
        if (alive) setLoading(false);
      });

    return () => {
      alive = false;
    };
  }, [refreshKey]);

  const refundByBooking = useMemo(() => {
    const map = new Map<string, Refund>();
    for (const refund of refunds) {
      if (!map.has(refund.booking_id)) map.set(refund.booking_id, refund);
    }
    return map;
  }, [refunds]);

  const today = todayISO();
  const eligibleBookings = useMemo(() => {
    return (bookings ?? [])
      .filter((booking) => {
        const isPaid = booking.status === "PAID" || booking.status === "CONFIRMED";
        if (!isPaid || booking.visit_date <= today) return false;
        const active = refundByBooking.get(booking.id);
        return !active || (active.status !== "PENDING" && active.status !== "APPROVED");
      })
      .sort((a, b) => a.visit_date.localeCompare(b.visit_date));
  }, [bookings, refundByBooking, today]);

  const pendingCount = refunds.filter((refund) => refund.status === "PENDING").length;
  const refundedAmount = refunds
    .filter((refund) => refund.status === "REFUNDED")
    .reduce((sum, refund) => sum + refund.amount, 0);

  return (
    <div>
      <PageHeader
        title="Refund"
        subtitle="Ajukan pengembalian dana atau pantau status pengajuan Anda"
      />

      <StatGrid cols={3}>
        <StatCard label="Booking Bisa Direfund" value={String(eligibleBookings.length)} />
        <StatCard label="Menunggu Diproses" value={String(pendingCount)} />
        <StatCard label="Total Dana Diterima" value={rupiah(refundedAmount)} accent />
      </StatGrid>

      {error && (
        <p className="mt-3 rounded-md bg-[#fdeaea] px-3 py-2 text-sm text-[#c53030]">{error}</p>
      )}

      <SectionCard
        title="Booking yang Bisa Direfund"
        subtitle="Refund pengunjung dikenakan potongan 10% dari total pembayaran."
        className="mt-3"
      >
        {loading && bookings === null ? (
          <div className="flex items-center gap-2 py-8 text-sm text-[#698097]">
            <Loader2 className="h-4 w-4 animate-spin" /> Memuat booking…
          </div>
        ) : eligibleBookings.length === 0 ? (
          <div className="py-8 text-center">
            <Banknote className="mx-auto h-10 w-10 text-[#94A3B8]" />
            <p className="mt-2 text-sm text-[#698097]">
              Tidak ada booking yang memenuhi syarat refund.
            </p>
            <p className="mx-auto mt-1 max-w-md text-xs leading-relaxed text-[#94A3B8]">
              Refund hanya tersedia untuk booking berstatus PAID/CONFIRMED dengan tanggal
              kunjungan minimal besok, dan belum memiliki pengajuan refund aktif.
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
              const breakdown = calculateRefundBreakdown(booking.total, "USER");
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

                  <div className="mt-2 space-y-1 border-t border-[#d7e3ed] pt-2 text-xs">
                    <div className="flex items-center justify-between gap-2">
                      <span className="text-[#698097]">Total pembayaran</span>
                      <span className="font-semibold tabular-nums text-[#17324d]">
                        {rupiah(breakdown.originalAmount)}
                      </span>
                    </div>
                    <div className="flex items-center justify-between gap-2 text-[#c53030]">
                      <span>Potongan {breakdown.deductionPercent}%</span>
                      <span className="font-semibold tabular-nums">
                        - {rupiah(breakdown.deductionAmount)}
                      </span>
                    </div>
                    <div className="flex items-center justify-between gap-2 border-t border-[#d7e3ed] pt-1">
                      <span className="font-semibold text-[#17324d]">Nominal diterima</span>
                      <span className="font-bold tabular-nums text-[#16876a]">
                        {rupiah(breakdown.netAmount)}
                      </span>
                    </div>
                  </div>

                  <button
                    type="button"
                    onClick={() => setSelected(booking)}
                    className="mt-3 inline-flex h-8 w-full items-center justify-center gap-1.5 rounded-md bg-[#1768ad] px-3 text-xs font-semibold text-white hover:bg-[#14548f]"
                  >
                    <Banknote className="h-3.5 w-3.5" /> Ajukan Refund
                  </button>
                </div>
              );
            })}
          </div>
        )}
      </SectionCard>

      <SectionCard
        title="Riwayat Refund"
        subtitle="Semua pengajuan refund Anda, termasuk refund yang diproses langsung oleh admin."
        className="mt-3"
      >
        <div className="panel-scroll rounded-md border border-[#e5edf3]">
          <table className="panel-table">
            <thead>
              <tr>
                <th>Kode Booking</th>
                <th>Kunjungan</th>
                <th>Jenis</th>
                <th>Total</th>
                <th>Potongan</th>
                <th>Diterima</th>
                <th>Tujuan</th>
                <th>Status</th>
                <th>Diajukan</th>
              </tr>
            </thead>
            <tbody>
              {loading ? (
                <tr>
                  <td colSpan={9} className="py-10 text-center text-sm text-[#698097]">
                    Memuat…
                  </td>
                </tr>
              ) : refunds.length === 0 ? (
                <tr>
                  <td colSpan={9} className="py-10 text-center text-sm text-[#94A3B8]">
                    Belum ada pengajuan refund.
                  </td>
                </tr>
              ) : (
                refunds.map((refund) => (
                  <tr key={refund.id}>
                    <td className="font-mono text-[11px] font-semibold text-[#17324d]">
                      {refund.booking_code}
                    </td>
                    <td className="tabular-nums">{refund.visit_date}</td>
                    <td>
                      <span className="inline-block rounded-full bg-[#eef4f8] px-2 py-0.5 text-[10px] font-bold text-[#29445e]">
                        {REFUND_INITIATOR_LABEL[refund.initiated_by] ?? refund.initiated_by}
                      </span>
                    </td>
                    <td className="tabular-nums">{rupiah(refund.original_amount)}</td>
                    <td className="tabular-nums text-[#c53030]">
                      {refund.deduction_amount > 0
                        ? `- ${rupiah(refund.deduction_amount)}`
                        : "Tanpa potongan"}
                    </td>
                    <td className="tabular-nums font-semibold text-[#17324d]">
                      {rupiah(refund.amount)}
                    </td>
                    <td>
                      {refund.bank_name ? (
                        <>
                          <p className="text-[11px] font-semibold text-[#17324d]">
                            {refund.bank_name}
                          </p>
                          <p className="text-[10px] text-[#698097]">
                            {REFUND_DESTINATION_LABEL[refund.destination_type] ?? "Rekening Bank"}
                            {refund.account_number ? ` · ${refund.account_number}` : ""}
                          </p>
                        </>
                      ) : (
                        <span className="text-[#94A3B8]">—</span>
                      )}
                    </td>
                    <td>
                      <span
                        className={`inline-block rounded-full px-2 py-0.5 text-[10px] font-bold ${
                          STATUS_CLASS[refund.status]
                        }`}
                      >
                        {REFUND_STATUS_LABEL[refund.status]}
                      </span>
                    </td>
                    <td className="text-[11px] text-[#698097]">
                      {formatDate(refund.created_at)}
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </SectionCard>

      <RefundRequestDialog
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
