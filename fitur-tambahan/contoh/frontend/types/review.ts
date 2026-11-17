export type ReviewStatus = "APPROVED" | "REJECTED";

export type ReviewSort = "newest" | "rating_desc" | "rating_asc";

export interface Review {
  id: string;
  user_id: string;
  booking_id?: string;
  destination_id?: string;
  rating: number;
  comment: string;
  status: ReviewStatus;
  is_approved: boolean;
  created_at: string;
  updated_at: string;
}

export interface ReviewSummary {
  average_rating: number;
  total_reviews: number;
}

export interface CreateReviewPayload {
  booking_id: string;
  destination_id?: string;
  rating: number;
  comment: string;
}

export interface UpdateReviewPayload {
  rating: number;
  comment: string;
}

export const REVIEW_STATUS_LABEL: Record<ReviewStatus, string> = {
  APPROVED: "Disetujui",
  REJECTED: "Ditolak",
};
