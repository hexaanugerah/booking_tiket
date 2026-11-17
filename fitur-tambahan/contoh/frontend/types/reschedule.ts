export interface Reschedule {
  id: string;
  booking_id: string;
  booking_code: string;
  old_visit_date: string;
  new_visit_date: string;
  reason: string;
  status: "APPROVED";
  created_at: string;
}

export interface CreateReschedulePayload {
  new_visit_date: string;
  reason: string;
}
