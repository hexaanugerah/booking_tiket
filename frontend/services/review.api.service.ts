import { apiClient } from "@/lib/api/client";
import type { ApiResponse } from "@/types/api";
import type { CreateReviewPayload, Review, UpdateReviewPayload } from "@/types/review";

export const reviewService = {
  /** Ulasan yang sudah disetujui admin — boleh dibaca tanpa login. */
  async listApproved(params: { page?: number; limit?: number } = {}): Promise<{ items: Review[]; total: number }> {
    const response = await apiClient.get<ApiResponse<Review[]>>("/reviews", {
      page: String(params.page ?? 1),
      limit: String(params.limit ?? 100),
    });
    return {
      items: response.data ?? [],
      total: response.meta?.total ?? response.data?.length ?? 0,
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
