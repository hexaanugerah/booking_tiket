import type { NextConfig } from "next";

// Origin backend untuk diteruskan server Next (rewrites). Hanya dipakai di sisi
// server, jadi tidak perlu prefix NEXT_PUBLIC_ dan tidak bocor ke browser.
const BACKEND_ORIGIN = process.env.BACKEND_ORIGIN ?? "http://localhost:8080";

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
  // Teruskan API dan berkas unggahan lewat origin frontend sendiri. Dengan ini
  // halaman yang dibuka dari HP, IP LAN, atau port forwarding VS Code tidak
  // perlu menjangkau port 8080 langsung (bebas CORS dan mixed content).
  async rewrites() {
    return [
      { source: "/api/v1/:path*", destination: `${BACKEND_ORIGIN}/api/v1/:path*` },
      { source: "/uploads/:path*", destination: `${BACKEND_ORIGIN}/uploads/:path*` },
    ];
  },
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
