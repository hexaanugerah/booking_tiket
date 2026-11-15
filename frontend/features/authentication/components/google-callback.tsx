"use client";

import { useEffect, useRef, useState } from "react";
import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { CheckCircle2, Loader2, XCircle } from "lucide-react";
import { authService } from "@/services/auth.api.service";
import { useAuthStore } from "@/store/auth.store";

// Komponen client callback login Google — backend mengembalikan pengguna ke sini
// membawa access_token, refresh_token, dan data user via query. Halaman ini
// menyimpan sesi lalu langsung menuntun ke beranda (atau /login bila gagal).
// Komponen client callback login Google. Dipanggil oleh server page
// app/(auth)/auth/google/callback/page.tsx di dalam <Suspense> karena
// memakai useSearchParams (wajib di Next.js App Router).
export function GoogleCallbackClient() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const user = useAuthStore((s) => s.user);
  const [error, setError] = useState<string | null>(null);
  const startedRef = useRef(false);

  useEffect(() => {
    // Guard React Strict Mode / remount: sesi hanya disimpan sekali.
    if (startedRef.current) return;
    startedRef.current = true;

    const err = searchParams.get("error");
    const accessToken = searchParams.get("access_token");
    const refreshToken = searchParams.get("refresh_token");
    const rawUser = searchParams.get("user");

    if (err) {
      setError(err);
      return;
    }
    if (!accessToken || !rawUser) {
      setError("Data login Google tidak diterima. Silakan coba lagi.");
      return;
    }

    try {
      const parsedUser = JSON.parse(rawUser);
      const { redirectTo } = authService.completeGoogleLogin({
        access_token: accessToken,
        refresh_token: refreshToken ?? undefined,
        user: parsedUser,
      });
      router.replace(redirectTo);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Gagal menyiapkan sesi Google.");
    }
  }, [searchParams, router]);

  // Sudah login (mis. callback dibuka ulang) — kembali ke beranda.
  useEffect(() => {
    if (user && !error) {
      const timer = setTimeout(() => router.replace("/"), 800);
      return () => clearTimeout(timer);
    }
  }, [user, error, router]);

  return (
    <main className="flex min-h-screen items-center justify-center bg-[#f3f8fc] px-4 py-10">
      <div className="w-full max-w-md rounded-2xl border border-[#d7e3ed] bg-white px-8 py-12 text-center shadow-sm">
        {error ? (
          <>
            <XCircle className="mx-auto h-12 w-12 text-red-500" />
            <h1 className="mt-4 font-heading text-2xl font-bold text-[#17324d]">
              Login Google Gagal
            </h1>
            <p className="mt-2 text-sm leading-relaxed text-[#698097]">{error}</p>
            <div className="mt-8 flex justify-center gap-3">
              <Link
                href="/login"
                className="inline-flex h-10 items-center rounded-md bg-[#1768ad] px-5 text-sm font-semibold text-white hover:bg-[#14548f]"
              >
                Coba Lagi
              </Link>
              <Link
                href="/"
                className="inline-flex h-10 items-center rounded-md border border-[#d7e3ed] bg-white px-5 text-sm font-semibold text-[#29445e] hover:bg-[#f2f7fb]"
              >
                Ke Beranda
              </Link>
            </div>
          </>
        ) : user ? (
          <>
            <CheckCircle2 className="mx-auto h-12 w-12 text-emerald-500" />
            <h1 className="mt-4 font-heading text-2xl font-bold text-[#17324d]">
              Login Berhasil
            </h1>
            <p className="mt-2 text-sm text-[#698097]">Mengalihkan ke beranda...</p>
          </>
        ) : (
          <>
            <Loader2 className="mx-auto h-12 w-12 animate-spin text-[#1768ad]" />
            <h1 className="mt-4 font-heading text-2xl font-bold text-[#17324d]">
              Menyelesaikan Login Google
            </h1>
            <p className="mt-2 text-sm text-[#698097]">Mohon tunggu sebentar...</p>
          </>
        )}
      </div>
    </main>
  );
}
