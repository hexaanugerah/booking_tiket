// Origin backend untuk route handler proxy (/api/v1 dan /uploads).
//
// Nilai ini dibaca SAAT REQUEST, bukan saat build. Dengan begitu URL backend
// bisa diganti lewat variabel/secret Worker tanpa membangun ulang frontend.
// Akses memakai tanda kurung siku supaya Next.js tidak mengganti ekspresi ini
// dengan nilai build-time.

export class BackendOriginError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "BackendOriginError";
  }
}

/**
 * Mengembalikan origin backend tanpa path, mis. https://api.kawahputih.id.
 * Melempar BackendOriginError bila konfigurasi belum diisi/tidak valid.
 */
export function resolveBackendOrigin(): string {
  const configured =
    process.env["BACKEND_ORIGIN"]?.trim() ||
    // next dev tetap bisa berjalan tanpa .env; Worker selalu production dan
    // wajib mengisi BACKEND_ORIGIN.
    (process.env.NODE_ENV !== "production" ? "http://localhost:8080" : "");

  if (!configured) {
    throw new BackendOriginError(
      "BACKEND_ORIGIN belum diatur. Isi dengan origin backend publik, mis. https://api.kawahputih.id"
    );
  }

  let url: URL;
  try {
    url = new URL(configured);
  } catch {
    throw new BackendOriginError(
      `BACKEND_ORIGIN bukan URL yang valid: ${configured}`
    );
  }

  if (url.protocol !== "http:" && url.protocol !== "https:") {
    throw new BackendOriginError(
      "BACKEND_ORIGIN harus memakai protokol http:// atau https://"
    );
  }

  // Hanya skema + host + port. Path diabaikan agar konfigurasi seperti
  // https://api.example.com/api/v1 tidak menghasilkan path ganda.
  return url.origin;
}
