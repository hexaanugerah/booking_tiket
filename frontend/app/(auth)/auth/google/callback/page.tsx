import type { Metadata } from "next";
import { Suspense } from "react";
import { GoogleCallbackClient } from "@/features/authentication/components/google-callback";

export const metadata: Metadata = { title: "Login Google" };

// Halaman callback login Google — tujuan redirect dari backend
// (GET /api/v1/auth/google/callback). Client dibungkus Suspense karena
// membaca query ?access_token / ?error lewat useSearchParams.
export default function GoogleCallbackPage() {
  return (
    <Suspense
      fallback={
        <main className="flex min-h-screen items-center justify-center bg-[#f3f8fc] px-4 py-10">
          <div className="w-full max-w-md rounded-2xl border border-[#d7e3ed] bg-white px-8 py-12 text-center shadow-sm">
            <h1 className="font-heading text-2xl font-bold text-[#17324d]">
              Menyelesaikan Login Google
            </h1>
            <p className="mt-2 text-sm text-[#698097]">Mohon tunggu sebentar...</p>
          </div>
        </main>
      }
    >
      <GoogleCallbackClient />
    </Suspense>
  );
}
