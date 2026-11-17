"use client";

import { useEffect, useState } from "react";
import {
  ArrowRight,
  CalendarClock,
  CheckCircle2,
  Eye,
  Loader2,
  XCircle,
} from "lucide-react";
import { toast } from "sonner";
import { PageHeader, SectionCard, StatCard, StatGrid } from "@/components/panel/panel-ui";
import { rescheduleService } from "@/services/reschedule.api.service";
import {
  RESCHEDULE_STATUS_LABEL,
  type Reschedule,
  type RescheduleStatus,
  type RescheduleSummary,
} from "@/types/reschedule";

const STATUS_CLASS: Record<RescheduleStatus, string> = {
  PENDING: "bg-[#fdf3e3] text-[#b7791f]",
  APPROVED: "bg-[#e6f7f0] text-[#16876a]",
  REJECTED: "bg-[#fdeaea] text-[#c53030]",
};

type RescheduleAction = "APPROVED" | "REJECTED";

const PAGE_SIZE = 10;

export default function AdminReschedulesPage() {
  const [rows, setRows] = useState<Reschedule[]>([]);
  const [total, setTotal] = useState(0);
  const [summary, setSummary] = useState<RescheduleSummary | null>(null);
  const [status, setStatus] = useState("");
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [refreshKey, setRefreshKey] = useState(0);

  const [detail, setDetail] = useState<Reschedule | null>(null);
  const [adminNote, setAdminNote] = useState("");
  const [submitting, setSubmitting] = useState<RescheduleAction | null>(null);

  useEffect(() => {
    let alive = true;
    setLoading(true);
    setError(null);

    Promise.all([
      rescheduleService.listAdmin({ page, limit: PAGE_SIZE, status: status || undefined }),
      rescheduleService.summary(),
    ])
      .then(([list, stat]) => {
        if (!alive) return;
        setRows(list.items);
        setTotal(list.total);
        setSummary(stat);
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
  }, [status, page, refreshKey]);

  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));

  function openDetail(item: Reschedule) {
    setDetail(item);
    setAdminNote(item.admin_note ?? "");
  }

  async function process(next: RescheduleAction) {
    if (!detail) return;

    if (
      next === "APPROVED" &&
      !confirm(
        `Setujui jadwal ulang ${detail.booking_code} dari ${detail.old_visit_date} ke ${detail.new_visit_date}? Tanggal booking dan e-ticket akan langsung diubah.`
      )
    ) {
      return;
    }

    setSubmitting(next);
    try {
      const updated = await rescheduleService.updateStatus(detail.id, {
        status: next,
        admin_note: adminNote,
      });
      setDetail(updated);
      setRefreshKey((key) => key + 1);
      toast.success(
        next === "APPROVED" ? "Jadwal ulang disetujui" : "Jadwal ulang ditolak",
        {
          description:
            next === "APPROVED"
              ? `${updated.booking_code} · ${updated.new_visit_date}`
              : updated.booking_code,
        }
      );
    } catch (err: any) {
      toast.error("Gagal memproses jadwal ulang", { description: err.message });
    } finally {
      setSubmitting(null);
    }
  }

  return (
    <div>
      <PageHeader
        title="Jadwal Ulang"
        subtitle="Tinjau pengajuan perubahan tanggal kunjungan dari pengunjung"
      />

      <StatGrid cols={3}>
        <StatCard label="Menunggu" value={String(summary?.pending ?? 0)} />
        <StatCard label="Disetujui" value={String(summary?.approved ?? 0)} />
        <StatCard label="Ditolak" value={String(summary?.rejected ?? 0)} accent />
      </StatGrid>

      <SectionCard
        title="Daftar Pengajuan"
        className="mt-3"
        action={<span className="text-sm font-semibold text-[#1768ad]">{total} data</span>}
      >
        <div className="mb-3 flex flex-wrap items-center gap-2">
          <select
            value={status}
            onChange={(event) => {
              setStatus(event.target.value);
              setPage(1);
            }}
            className="h-9 rounded-md border border-[#d7e3ed] bg-white px-3 text-sm text-[#29445e] outline-none focus:border-[#1768ad]"
            aria-label="Filter status jadwal ulang"
          >
            <option value="">Semua status</option>
            {(Object.keys(RESCHEDULE_STATUS_LABEL) as RescheduleStatus[]).map((value) => (
              <option key={value} value={value}>
                {RESCHEDULE_STATUS_LABEL[value]}
              </option>
            ))}
          </select>
        </div>

        {error && (
          <p className="mb-3 rounded-md bg-[#fdeaea] px-3 py-2 text-sm text-[#c53030]">{error}</p>
        )}

        <div className="panel-scroll rounded-md border border-[#e5edf3]">
          <table className="panel-table">
            <thead>
              <tr>
                <th>No</th>
                <th>Booking</th>
                <th>Pengunjung</th>
                <th>Tanggal</th>
                <th>Alasan</th>
                <th>Status</th>
                <th>Diajukan</th>
                <th>Aksi</th>
              </tr>
            </thead>
            <tbody>
              {loading ? (
                <tr>
                  <td colSpan={8} className="py-10 text-center text-sm text-[#698097]">
                    Memuat…
                  </td>
                </tr>
              ) : rows.length === 0 ? (
                <tr>
                  <td colSpan={8} className="py-10 text-center text-sm text-[#94A3B8]">
                    Belum ada pengajuan jadwal ulang.
                  </td>
                </tr>
              ) : (
                rows.map((item, index) => (
                  <tr key={item.id}>
                    <td className="tabular-nums text-[#698097]">
                      {(page - 1) * PAGE_SIZE + index + 1}.
                    </td>
                    <td>
                      <p className="font-mono text-[11px] font-semibold text-[#17324d]">
                        {item.booking_code}
                      </p>
                      <p className="text-[10px] text-[#698097]">{item.booking_status}</p>
                    </td>
                    <td>
                      <p className="font-medium text-[#17324d]">{item.visitor_name}</p>
                      <p className="text-[10px] text-[#698097]">{item.email}</p>
                    </td>
                    <td className="whitespace-nowrap text-[11px]">
                      <span className="tabular-nums text-[#698097]">{item.old_visit_date}</span>
                      <span className="mx-1 inline-flex text-[#94A3B8]">
                        <ArrowRight className="h-3 w-3" />
                      </span>
                      <span className="tabular-nums font-semibold text-[#17324d]">
                        {item.new_visit_date}
                      </span>
                    </td>
                    <td className="max-w-[220px]">
                      <p className="line-clamp-2 text-[11px] text-[#29445e]">{item.reason}</p>
                    </td>
                    <td>
                      <span
                        className={`inline-block rounded-full px-2 py-0.5 text-[10px] font-bold ${STATUS_CLASS[item.status]}`}
                      >
                        {RESCHEDULE_STATUS_LABEL[item.status]}
                      </span>
                    </td>
                    <td className="text-[11px] text-[#698097]">
                      {formatDate(item.created_at)}
                    </td>
                    <td>
                      <button
                        type="button"
                        onClick={() => openDetail(item)}
                        className="grid h-7 w-7 place-items-center rounded-md border border-[#d7e3ed] text-[#29445e] hover:bg-[#f2f7fb]"
                        aria-label={`Detail jadwal ulang ${item.booking_code}`}
                      >
                        <Eye className="h-3.5 w-3.5" />
                      </button>
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>

        <div className="mt-3 flex items-center justify-end gap-1 text-xs text-[#698097]">
          <button
            type="button"
            onClick={() => setPage((current) => Math.max(1, current - 1))}
            disabled={page <= 1}
            className="rounded-md border border-[#d7e3ed] px-2 py-1 disabled:opacity-40"
          >
            ‹
          </button>
          <span className="px-2">
            Hal {page} / {totalPages}
          </span>
          <button
            type="button"
            onClick={() => setPage((current) => Math.min(totalPages, current + 1))}
            disabled={page >= totalPages}
            className="rounded-md border border-[#d7e3ed] px-2 py-1 disabled:opacity-40"
          >
            ›
          </button>
        </div>
      </SectionCard>

      {detail && (
        <div
          className="fixed inset-0 z-50 grid place-items-center bg-black/40 p-4"
          role="dialog"
          aria-modal="true"
          aria-labelledby="reschedule-detail-title"
          onClick={() => setDetail(null)}
        >
          <div
            className="max-h-[92vh] w-full max-w-lg overflow-y-auto rounded-md bg-white p-5"
            onClick={(event) => event.stopPropagation()}
          >
            <div className="flex items-start justify-between gap-3 border-b border-[#e5edf3] pb-3">
              <div className="flex items-center gap-2">
                <span className="grid h-9 w-9 place-items-center rounded-full bg-[#e9f3fb] text-[#1768ad]">
                  <CalendarClock className="h-4 w-4" aria-hidden="true" />
                </span>
                <div>
                  <h3
                    id="reschedule-detail-title"
                    className="text-base font-bold text-[#17324d]"
                  >
                    Detail Jadwal Ulang
                  </h3>
                  <p className="text-[11px] text-[#698097]">{detail.booking_code}</p>
                </div>
              </div>
              <span
                className={`inline-block rounded-full px-2 py-0.5 text-[10px] font-bold ${STATUS_CLASS[detail.status]}`}
              >
                {RESCHEDULE_STATUS_LABEL[detail.status]}
              </span>
            </div>

            <dl className="mt-3 space-y-2 text-sm">
              <Row label="Pengunjung" value={detail.visitor_name} />
              <Row label="Email" value={detail.email} />
              <Row label="Tanggal lama" value={detail.old_visit_date} />
              <Row label="Tanggal baru" value={detail.new_visit_date} />
              <Row label="Status booking" value={detail.booking_status} />
              <Row label="Diajukan" value={formatDateTime(detail.created_at)} />
              {detail.processed_at && (
                <Row label="Diproses" value={formatDateTime(detail.processed_at)} />
              )}
            </dl>

            <div className="mt-3 rounded-md bg-[#f7fbff] px-3 py-2 text-xs text-[#29445e]">
              <p className="font-semibold text-[#17324d]">Alasan pengunjung</p>
              <p className="mt-1 leading-relaxed">{detail.reason}</p>
            </div>

            {detail.admin_note && (
              <div className="mt-2 rounded-md bg-[#fdf3e3] px-3 py-2 text-xs text-[#8a6420]">
                <p className="font-semibold">Catatan admin</p>
                <p className="mt-1 leading-relaxed">{detail.admin_note}</p>
              </div>
            )}

            {detail.status === "PENDING" && (
              <label className="mt-3 block">
                <span className="text-xs font-semibold text-[#29445e]">
                  Catatan admin (opsional)
                </span>
                <textarea
                  value={adminNote}
                  onChange={(event) => setAdminNote(event.target.value)}
                  maxLength={500}
                  rows={2}
                  placeholder="Contoh: Kuota tanggal baru tersedia, silakan datang sesuai jadwal."
                  className="mt-1 w-full rounded-md border border-[#d7e3ed] px-3 py-2 text-sm text-[#17324d] outline-none placeholder:text-[#94A3B8] focus:border-[#1768ad]"
                />
              </label>
            )}

            <div className="mt-4 flex flex-wrap justify-end gap-2">
              <button
                type="button"
                onClick={() => setDetail(null)}
                className="h-9 rounded-md border border-[#d7e3ed] px-4 text-sm font-semibold text-[#29445e] hover:bg-[#f2f7fb]"
              >
                Tutup
              </button>

              {detail.status === "PENDING" && (
                <>
                  <button
                    type="button"
                    onClick={() => process("REJECTED")}
                    disabled={submitting !== null}
                    className="inline-flex h-9 items-center gap-1.5 rounded-md bg-[#c53030] px-4 text-sm font-semibold text-white hover:bg-[#9b2626] disabled:opacity-60"
                  >
                    {submitting === "REJECTED" ? (
                      <Loader2 className="h-3.5 w-3.5 animate-spin" />
                    ) : (
                      <XCircle className="h-3.5 w-3.5" />
                    )}
                    Tolak
                  </button>
                  <button
                    type="button"
                    onClick={() => process("APPROVED")}
                    disabled={submitting !== null}
                    className="inline-flex h-9 items-center gap-1.5 rounded-md bg-[#16876a] px-4 text-sm font-semibold text-white hover:bg-[#126e57] disabled:opacity-60"
                  >
                    {submitting === "APPROVED" ? (
                      <Loader2 className="h-3.5 w-3.5 animate-spin" />
                    ) : (
                      <CheckCircle2 className="h-3.5 w-3.5" />
                    )}
                    Setujui
                  </button>
                </>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

function Row({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-start justify-between gap-4">
      <dt className="shrink-0 text-[#698097]">{label}</dt>
      <dd className="break-all text-right font-medium text-[#17324d]">{value}</dd>
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

function formatDateTime(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return date.toLocaleString("id-ID", {
    day: "2-digit",
    month: "short",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}
