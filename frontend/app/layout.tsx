import type { Metadata } from "next";
import type { ReactNode } from "react";
import "./globals.css";
import { QueryProvider } from "@/providers/query-provider";
import { ThemeProvider } from "@/providers/theme-provider";
import { ToastProvider } from "@/providers/toast-provider";
import { AuthProvider } from "@/providers/auth-provider";

export const metadata: Metadata = {
  title: {
    default: "Kawah Putih Rancabali — Wisata Alam Bandung",
    template: "%s | Kawah Putih Rancabali",
  },
  description:
    "Pesan tiket wisata Kawah Putih Rancabali secara online. Informasi lengkap, harga tiket, galeri, dan artikel wisata terbaru.",
  keywords: ["kawah putih", "rancabali", "wisata bandung", "tiket wisata online"],
  openGraph: {
    type: "website",
    locale: "id_ID",
    siteName: "Kawah Putih Rancabali",
  },
  twitter: { card: "summary_large_image" },
  robots: { index: true, follow: true },
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="id" suppressHydrationWarning>
      <body>
        <QueryProvider>
          <ThemeProvider>
            <AuthProvider>
              {children}
              <ToastProvider />
            </AuthProvider>
          </ThemeProvider>
        </QueryProvider>
      </body>
    </html>
  );
}
