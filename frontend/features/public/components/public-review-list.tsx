"use client";

import { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { ArrowRight } from "lucide-react";
import { reviewService } from "@/services/review.api.service";
import type { Review } from "@/types/review";
import { RatingStars, ReviewCard } from "@/features/review/components/review-display";

export function PublicReviewList() {
  const [reviews, setReviews] = useState<Review[] | null>(null);
  const [total, setTotal] = useState(0);
  const [error, setError] = useState(false);

  useEffect(() => {
    let alive = true;
    reviewService
      .listApproved({ limit: 100 })
      .then(({ items, total: count }) => {
        if (!alive) return;
        setReviews(items);
        setTotal(count);
      })
      .catch(() => {
        if (!alive) return;
        setReviews([]);
        setError(true);
      });
    return () => {
      alive = false;
    };
  }, []);

  const averageRating = useMemo(() => {
    if (!reviews || reviews.length === 0) return 0;
    return reviews.reduce((sum, review) => sum + review.rating, 0) / reviews.length;
  }, [reviews]);

  return (
    <div>
      <div className="flex flex-wrap items-end justify-between gap-4">
        <div className="max-w-2xl">
          <p className="text-xs font-bold uppercase tracking-wide text-[#1768ad]">Ulasan Pengunjung</p>
          <h1 className="mt-2 font-heading text-3xl font-bold text-[#17324d] sm:text-4xl">
            Apa kata pengunjung Kawah Putih
          </h1>
          <p className="mt-3 text-[15px] leading-relaxed text-[#698097]">
            Rating dan ulasan hanya berasal dari pengunjung dengan booking yang sudah dibayar dan disetujui
            admin.
          </p>
        </div>

        {reviews !== null && reviews.length > 0 && (
          <div className="flex items-center gap-4 rounded-md border border-[#d7e3ed] bg-white px-5 py-4">
            <span className="text-4xl font-bold leading-none text-[#17324d]">{averageRating.toFixed(1)}</span>
            <div>
              <RatingStars value={Math.round(averageRating)} />
              <p className="mt-1 text-xs text-[#698097]">{total} ulasan</p>
            </div>
          </div>
        )}
      </div>

      {reviews === null ? (
        <div className="mt-8 grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
          {Array.from({ length: 6 }, (_, i) => (
            <div key={i} className="h-44 animate-pulse rounded-md border border-[#d7e3ed] bg-white" />
          ))}
        </div>
      ) : error ? (
        <p className="mt-8 rounded-md border border-[#d7e3ed] bg-white px-5 py-4 text-sm text-[#698097]">
          Ulasan belum dapat dimuat. Silakan coba muat ulang halaman.
        </p>
      ) : reviews.length === 0 ? (
        <p className="mt-8 rounded-md border border-[#d7e3ed] bg-white px-5 py-4 text-sm text-[#698097]">
          Belum ada ulasan yang tayang. Bagikan pengalamanmu dan jadilah ulasan pertama setelah disetujui
          admin.
        </p>
      ) : (
        <div className="mt-8 grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
          {reviews.map((review) => (
            <ReviewCard key={review.id} review={review} className="bg-white" />
          ))}
        </div>
      )}

      <div className="mt-8 flex flex-wrap items-center justify-between gap-4 rounded-lg border border-[#1768ad]/20 bg-[#f7fbff] p-5">
        <div>
          <p className="text-sm font-bold text-[#17324d]">Sudah pernah berkunjung?</p>
          <p className="mt-1 text-xs text-[#698097]">
            Beri rating 1-5 dan tulis ulasan langsung dari halaman utama.
          </p>
        </div>
        <Link
          href="/#ulasan"
          className="inline-flex h-10 items-center gap-2 rounded-md bg-[#1768ad] px-5 text-sm font-semibold text-white hover:bg-[#14548f]"
        >
          Tulis Ulasan <ArrowRight className="h-4 w-4" aria-hidden="true" />
        </Link>
      </div>
    </div>
  );
}
