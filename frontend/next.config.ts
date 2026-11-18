import type { NextConfig } from "next";

// Proxy /api/v1 dan /uploads TIDAK memakai rewrites di sini karena target
// rewrites dibekukan saat build. Proxy ditangani route handler runtime
// (app/api/v1/[...path]/route.ts dan app/uploads/[...path]/route.ts) yang
// membaca BACKEND_ORIGIN pada setiap request.

const nextConfig: NextConfig = {
  reactStrictMode: true,
  images: {
    // URL gambar galeri admin memasukkan host/protokol apa pun, jadi izinkan
    // http & https dari host mana pun. Di produksi sebaiknya pakai HTTPS.
    remotePatterns: [
      { protocol: "https", hostname: "**" },
      { protocol: "http", hostname: "**" },
    ],
  },
  // Teruskan API dan berkas unggahan lewat origin frontend sendiri. Proxy-nya
  // ada di route handler runtime, bukan rewrites build-time, supaya origin
  // backend dapat diganti tanpa membangun ulang Worker.
  async headers() {
    return [
      {
        source: "/:path*",
        headers: [
          { key: "X-Frame-Options", value: "DENY" },
          { key: "X-Content-Type-Options", value: "nosniff" },
          { key: "Referrer-Policy", value: "strict-origin-when-cross-origin" },
        ],
      },
    ];
  },
};

export default nextConfig;
