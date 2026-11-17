import { apiClient } from "@/lib/api/client";
import type { ApiResponse } from "@/types/api";
import type { CreateReschedulePayload, Reschedule } from "@/types/reschedule";

export const rescheduleService = {
  /** Ubah tanggal kunjungan booking milik pengunjung. */
  async create(bookingId: string, payload: CreateReschedulePayload): Promise<Reschedule> {
    const response = await apiClient.post<ApiResponse<Reschedule>>(
      `/bookings/${bookingId}/reschedule`,
      payload
    );
    if (!response.success || !response.data) {
      throw new Error(response.message || "Gagal mengubah jadwal kunjungan");
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
};
