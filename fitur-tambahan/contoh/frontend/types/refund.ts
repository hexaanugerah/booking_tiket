export type RefundStatus = "PENDING" | "APPROVED" | "REJECTED" | "REFUNDED";

export const REFUND_STATUS_LABEL: Record<RefundStatus, string> = {
  PENDING: "Menunggu",
  APPROVED: "Disetujui",
  REJECTED: "Ditolak",
  REFUNDED: "Dana Dikirim",
};

export interface Refund {
  id: string;
  booking_id: string;
  user_id: string;
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
  bank_name: string;
  account_name: string;
  account_number: string;
}

export interface UpdateRefundStatusPayload {
  status: "APPROVED" | "REJECTED" | "REFUNDED";
  admin_note?: string;
}
