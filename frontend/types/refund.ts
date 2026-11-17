export type RefundStatus = "PENDING" | "APPROVED" | "REJECTED" | "REFUNDED";

export type RefundInitiatedBy = "USER" | "ADMIN";

/** Tujuan pengembalian dana: rekening bank atau e-wallet. */
export type RefundDestinationType = "BANK" | "EWALLET";

/** Potongan refund untuk pengajuan pengunjung. Refund langsung admin: 0%. */
export const USER_REFUND_DEDUCTION_PERCENT = 10;

export const REFUND_INITIATOR_LABEL: Record<RefundInitiatedBy, string> = {
  USER: "Pengunjung",
  ADMIN: "Admin",
};

export const REFUND_STATUS_LABEL: Record<RefundStatus, string> = {
  PENDING: "Menunggu",
  APPROVED: "Disetujui",
  REJECTED: "Ditolak",
  REFUNDED: "Dana Dikirim",
};

export const REFUND_DESTINATION_LABEL: Record<RefundDestinationType, string> = {
  BANK: "Rekening Bank",
  EWALLET: "E-Wallet",
};

/** Pilihan bank tujuan refund. */
export const REFUND_BANK_OPTIONS = [
  "BCA",
  "BRI",
  "BNI",
  "Mandiri",
  "CIMB Niaga",
  "Permata",
  "Danamon",
  "BTN",
  "Bank BJB",
  "Bank Sinarmas",
  "SeaBank",
  "Bank Jago",
] as const;

/** Pilihan e-wallet tujuan refund. */
export const REFUND_EWALLET_OPTIONS = [
  "DANA",
  "OVO",
  "GoPay",
  "ShopeePay",
  "LinkAja",
  "Jenius",
  "iSaku",
] as const;

/** Label nomor tujuan yang berubah sesuai jenis tujuan refund. */
export function refundAccountNumberLabel(type: RefundDestinationType): string {
  return type === "EWALLET" ? "Nomor e-wallet / HP" : "Nomor rekening";
}

/** Label nama pemilik yang berubah sesuai jenis tujuan refund. */
export function refundAccountNameLabel(type: RefundDestinationType): string {
  return type === "EWALLET" ? "Nama pemilik e-wallet" : "Nama pemilik rekening";
}

export interface Refund {
  id: string;
  booking_id: string;
  user_id: string;
  initiated_by: RefundInitiatedBy;
  destination_type: RefundDestinationType;
  original_amount: number;
  deduction_percent: number;
  deduction_amount: number;
  /** Nominal bersih yang dikembalikan ke pengunjung. */
  amount: number;
  reason: string;
  bank_name: string;
  account_name: string;
  account_number: string;
  status: RefundStatus;
  admin_note?: string;
  processed_by?: string;
  processed_at?: string;

  // Data booking dari join backend.
  booking_code: string;
  visitor_name: string;
  email: string;
  visit_date: string;
  booking_status: string;

  created_at: string;
  updated_at: string;
}

export interface RefundSummary {
  pending: number;
  approved: number;
  refunded: number;
  refunded_amount: number;
}

export interface CreateRefundPayload {
  booking_id: string;
  reason: string;
  destination_type: RefundDestinationType;
  bank_name: string;
  account_name: string;
  account_number: string;
}

/** Payload refund langsung oleh admin (tanpa potongan). */
export interface CreateAdminRefundPayload {
  booking_id: string;
  reason?: string;
  destination_type?: RefundDestinationType;
  bank_name?: string;
  account_name?: string;
  account_number?: string;
  admin_note?: string;
}

export interface UpdateRefundStatusPayload {
  status: "APPROVED" | "REJECTED" | "REFUNDED";
  admin_note?: string;
}
