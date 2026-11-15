// Helper untuk URL media (berkas unggahan) yang disajikan backend di /uploads.

import { resolveApiBaseUrl } from "@/lib/api-base";

// Origin backend: base URL API tanpa akhiran /api/v1. Memakai resolveApiBaseUrl
// agar tetap benar saat halaman dibuka dari HP/PC lain di jaringan yang sama.
export const backendOrigin = resolveApiBaseUrl().replace(/\/api\/v1\/?$/, "");

/**
 * Mengubah path media relatif (mis. "/uploads/gallery/x.jpg") menjadi URL
 * absolut ke backend. URL absolut dilewatkan apa adanya.
 */
export function mediaUrl(path?: string | null): string | undefined {
  if (!path) return undefined;
  if (/^https?:\/\//i.test(path)) return path;
  if (path.startsWith("/")) return `${backendOrigin}${path}`;
  return path;
}
