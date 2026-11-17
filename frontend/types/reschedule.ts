export type RescheduleStatus = "PENDING" | "APPROVED" | "REJECTED";

export const RESCHEDULE_STATUS_LABEL: Record<RescheduleStatus, string> = {
  PENDING: "Menunggu Persetujuan",
  APPROVED: "Disetujui",
  REJECTED: "Ditolak",
};

export interface Reschedule {
  id: string;
  booking_id: string;
  booking_code: string;
  user_id: string;
  old_visit_date: string;
  new_visit_date: string;
  reason: string;
  status: RescheduleStatus;
  admin_note?: string;
  processed_by?: string;
  processed_at?: string;

  // Data booking dari join backend.
  visitor_name: string;
  email: string;
  visit_date: string;
  booking_status: string;

  created_at: string;
  updated_at: string;
}

/** Batas maksimal jadwal ulang per booking (sama dengan aturan backend). */
export const MAX_RESCHEDULES_PER_BOOKING = 2;

/** Status yang masih memakai jatah jadwal ulang (pengajuan ditolak tidak). */
export const ACTIVE_RESCHEDULE_STATUSES: RescheduleStatus[] = ["PENDING", "APPROVED"];

export interface CreateReschedulePayload {
  new_visit_date: string;
  reason: string;
}

export interface UpdateRescheduleStatusPayload {
  status: "APPROVED" | "REJECTED";
  admin_note?: string;
}

export interface RescheduleSummary {
  pending: number;
  approved: number;
  rejected: number;
}
