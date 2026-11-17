import { apiClient } from "@/lib/api/client";
import type { ApiResponse } from "@/types/api";
import type {
  CreateReschedulePayload,
  Reschedule,
  RescheduleSummary,
  UpdateRescheduleStatusPayload,
} from "@/types/reschedule";

export interface RescheduleListParams {
  page?: number;
  limit?: number;
  status?: string;
}

export const rescheduleService = {
  /** Mengajukan jadwal ulang booking milik pengunjung (menunggu admin). */
  async create(bookingId: string, payload: CreateReschedulePayload): Promise<Reschedule> {
    const response = await apiClient.post<ApiResponse<Reschedule>>(
      `/bookings/${bookingId}/reschedule`,
      payload
    );
    if (!response.success || !response.data) {
      throw new Error(response.message || "Gagal mengajukan jadwal ulang");
    }
    return response.data;
  },

  /** Riwayat jadwal ulang untuk satu booking. */
  async list(bookingId: string): Promise<Reschedule[]> {
    const response = await apiClient.get<ApiResponse<Reschedule[]>>(
      `/bookings/${bookingId}/reschedules`
    );
    return response.data ?? [];
  },

  /** Seluruh riwayat jadwal ulang milik pengunjung yang sedang login. */
  async listMine(): Promise<Reschedule[]> {
    const response = await apiClient.get<ApiResponse<Reschedule[]>>("/reschedules/mine");
    return response.data ?? [];
  },

  /** Daftar pengajuan jadwal ulang untuk admin. */
  async listAdmin(params: RescheduleListParams = {}): Promise<{ items: Reschedule[]; total: number }> {
    const response = await apiClient.get<ApiResponse<Reschedule[]>>("/admin/reschedules", {
      page: String(params.page ?? 1),
      limit: String(params.limit ?? 10),
      status: params.status ?? "",
    });
    return {
      items: response.data ?? [],
      total: response.meta?.total ?? response.data?.length ?? 0,
    };
  },

  async summary(): Promise<RescheduleSummary> {
    const response = await apiClient.get<ApiResponse<RescheduleSummary>>(
      "/admin/reschedules/summary"
    );
    if (!response.success || !response.data) {
      throw new Error(response.message || "Gagal memuat ringkasan jadwal ulang");
    }
    return response.data;
  },

  /** Menyetujui atau menolak pengajuan jadwal ulang. */
  async updateStatus(id: string, payload: UpdateRescheduleStatusPayload): Promise<Reschedule> {
    const response = await apiClient.patch<ApiResponse<Reschedule>>(
      `/admin/reschedules/${id}/status`,
      payload
    );
    if (!response.success || !response.data) {
      throw new Error(response.message || "Gagal memperbarui status jadwal ulang");
    }
    return response.data;
  },
};
