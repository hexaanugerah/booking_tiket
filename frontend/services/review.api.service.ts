import { apiClient } from "@/lib/api/client";
import type { ApiResponse } from "@/types/api";
import type {
  CreateReviewPayload,
  Review,
  ReviewSort,
  ReviewSummary,
  UpdateReviewPayload,
} from "@/types/review";

export interface ReviewListParams {
  page?: number;
  limit?: number;
  rating?: number;
  sort?: ReviewSort;
}

export const reviewService = {
  /** Ulasan yang sudah tayang — boleh dibaca tanpa login. */
  async listApproved(
    params: ReviewListParams = {}
  ): Promise<{ items: Review[]; total: number }> {
    const response = await apiClient.get<ApiResponse<Review[]>>("/reviews", {
      page: String(params.page ?? 1),
      limit: String(params.limit ?? 100),
      rating: params.rating ? String(params.rating) : "",
      sort: params.sort ?? "newest",
    });
    return {
      items: response.data ?? [],
      total: response.meta?.total ?? response.data?.length ?? 0,
    };
  },

  /** Rata-rata rating dan jumlah ulasan tayang dari seluruh data. */
  async summary(): Promise<ReviewSummary> {
    const response = await apiClient.get<ApiResponse<ReviewSummary>>("/reviews/summary");
    if (!response.success || !response.data) {
      throw new Error(response.message || "Gagal memuat ringkasan ulasan");
    }
    return {
      average_rating: Number(response.data.average_rating) || 0,
      total_reviews: Number(response.data.total_reviews) || 0,
    };
  },

  async listMine(): Promise<Review[]> {
    const response = await apiClient.get<ApiResponse<Review[]>>("/reviews/mine", { limit: "100" });
    return response.data ?? [];
  },

  async create(payload: CreateReviewPayload): Promise<Review> {
    const response = await apiClient.post<ApiResponse<Review>>("/reviews", payload);
    if (!response.success || !response.data) {
      throw new Error(response.message || "Gagal mengirim ulasan");
    }
    return response.data;
  },

  async update(id: string, payload: UpdateReviewPayload): Promise<Review> {
    const response = await apiClient.put<ApiResponse<Review>>(`/reviews/${id}`, payload);
    if (!response.success || !response.data) {
      throw new Error(response.message || "Gagal memperbarui ulasan");
    }
    return response.data;
  },

  async remove(id: string): Promise<void> {
    await apiClient.delete<ApiResponse<null>>(`/reviews/${id}`);
  },
};
