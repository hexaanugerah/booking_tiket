import { apiClient } from "@/lib/api/client";
import type { ApiResponse } from "@/types/api";
import type {
  CreateRefundPayload,
  Refund,
  RefundSummary,
  UpdateRefundStatusPayload,
} from "@/types/refund";

export interface RefundListParams {
  page?: number;
  limit?: number;
  status?: string;
}

export const refundService = {
  /** Pengunjung mengajukan refund untuk booking miliknya. */
  async create(payload: CreateRefundPayload): Promise<Refund> {
    const response = await apiClient.post<ApiResponse<Refund>>("/refunds", payload);
    if (!response.success || !response.data) {
      throw new Error(response.message || "Gagal mengirim pengajuan refund");
    }
    return response.data;
  },

  /** Riwayat refund milik pengunjung yang sedang login. */
  async listMine(): Promise<Refund[]> {
    const response = await apiClient.get<ApiResponse<Refund[]>>("/refunds/mine", {
      limit: "100",
    });
    return response.data ?? [];
  },

  /** Daftar refund untuk admin. */
  async list(params: RefundListParams = {}): Promise<{ items: Refund[]; total: number }> {
    const response = await apiClient.get<ApiResponse<Refund[]>>("/admin/refunds", {
      page: String(params.page ?? 1),
      limit: String(params.limit ?? 10),
      status: params.status ?? "",
    });
    return {
      items: response.data ?? [],
      total: response.meta?.total ?? response.data?.length ?? 0,
    };
  },

  async summary(): Promise<RefundSummary> {
    const response = await apiClient.get<ApiResponse<RefundSummary>>("/admin/refunds/summary");
    if (!response.success || !response.data) {
      throw new Error(response.message || "Gagal memuat ringkasan refund");
    }
    return response.data;
  },

  async updateStatus(id: string, payload: UpdateRefundStatusPayload): Promise<Refund> {
    const response = await apiClient.patch<ApiResponse<Refund>>(
      `/admin/refunds/${id}/status`,
      payload
    );
    if (!response.success || !response.data) {
      throw new Error(response.message || "Gagal memperbarui status refund");
    }
    return response.data;
  },
};
