// Base URL API yang tetap benar saat aplikasi dibuka dari perangkat lain.
//
// NEXT_PUBLIC_* di-inline saat build, jadi kalau isinya
// "http://localhost:8080/api/v1" maka dari HP alamat itu menunjuk ke HP itu
// sendiri. Helper ini mengganti host loopback dengan hostname halaman yang
// sedang dibuka (mis. 192.168.1.5) sehingga satu build bisa dipakai dari PC
// maupun HP di jaringan yang sama.

const ENV_API_BASE_URL =
  process.env.NEXT_PUBLIC_API_BASE_URL ??
  process.env.NEXT_PUBLIC_API_URL ??
  "http://localhost:8080/api/v1";

function isLoopback(hostname: string): boolean {
  return hostname === "localhost" || hostname === "127.0.0.1" || hostname === "[::1]";
}

// Alamat privat LAN (RFC1918) — host ini masih menunjuk ke komputer yang sama
// dengan backend saat pengembangan.
function isPrivateHost(hostname: string): boolean {
  return (
    isLoopback(hostname) ||
    /^10(?:\.\d{1,3}){3}$/.test(hostname) ||
    /^192\.168(?:\.\d{1,3}){2}$/.test(hostname) ||
    /^172\.(?:1[6-9]|2\d|3[01])(?:\.\d{1,3}){2}$/.test(hostname)
  );
}

export function resolveApiBaseUrl(): string {
  if (typeof window === "undefined") return ENV_API_BASE_URL;

  try {		const url = new URL(ENV_API_BASE_URL);
		// Hanya ganti host kalau halaman dibuka dari loopback/IP LAN. Untuk domain
		// hasil port forwarding/tunnel (mis. *.devtunnels.ms atau *.trycloudflare.com)
		// pakai jalur relatif: server Next meneruskannya lewat rewrites di
		// next.config.ts, jadi backend hanya perlu dijangkau dari mesin ini.
		if (isPrivateHost(url.hostname)) {
			// Halaman dibuka dari loopback/IP LAN: arahkan langsung ke host yang sama.
			if (isPrivateHost(window.location.hostname)) {
				url.hostname = window.location.hostname;
				return url.toString().replace(/\/$/, "");
			}
			// Halaman publik (deployment Workers/tunnel): pakai jalur relatif supaya
			// server yang meneruskan ke backend. Ini menghindari mixed content dan
			// CORS karena browser hanya bicara ke origin halaman itu sendiri.
			return url.pathname.replace(/\/$/, "") || "/api/v1";
		}
  } catch {
    // Nilai env bukan URL absolut (mis. path relatif) — pakai apa adanya.
  }
  return ENV_API_BASE_URL;
}

export const API_BASE_URL = resolveApiBaseUrl();
