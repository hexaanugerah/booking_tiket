"use client";

import { useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { ArrowRight, Star } from "lucide-react";
import { reviewService } from "@/services/review.api.service";
import type { Review, ReviewSort, ReviewSummary } from "@/types/review";
import { RatingStars, ReviewCard } from "@/features/review/components/review-display";
import { cn } from "@/lib/utils";

const SORT_OPTIONS: { value: ReviewSort; label: string }[] = [
  { value: "newest", label: "Terbaru" },
  { value: "rating_desc", label: "Rating Tertinggi" },
  { value: "rating_asc", label: "Rating Terendah" },
];

const SORT_LABEL: Record<ReviewSort, string> = {
  newest: "Terbaru",
  rating_desc: "Rating Tertinggi",
  rating_asc: "Rating Terendah",
};

const RATING_OPTIONS = [5, 4, 3, 2, 1] as const;

export function PublicReviewList() {
  const [reviews, setReviews] = useState<Review[] | null>(null);
  const [total, setTotal] = useState(0);
  const [summary, setSummary] = useState<ReviewSummary | null>(null);
  const [error, setError] = useState(false);
  const [rating, setRating] = useState<number | null>(null);
  const [sort, setSort] = useState<ReviewSort>("newest");

  useEffect(() => {
    let alive = true;
    reviewService
      .summary()
      .then((data) => {
        if (alive) setSummary(data);
      })
      .catch(() => {});
    return () => {
      alive = false;
    };
  }, []);

  useEffect(() => {
    let alive = true;
    setReviews(null);
    setError(false);

    reviewService
      .listApproved({
        limit: 100,
        rating: rating ?? undefined,
        sort,
      })
      .then(({ items, total: count }) => {
        if (!alive) return;
        setReviews(items);
        setTotal(count);
      })
      .catch(() => {
        if (!alive) return;
        setReviews([]);
        setTotal(0);
        setError(true);
      });

    return () => {
      alive = false;
    };
  }, [rating, sort]);

  const averageRating = useMemo(() => {
    if (summary && summary.total_reviews > 0) return summary.average_rating;
    if (!reviews || reviews.length === 0) return 0;
    return reviews.reduce((sum, review) => sum + review.rating, 0) / reviews.length;
  }, [summary, reviews]);

  const displayTotal = summary?.total_reviews ?? total;

  return (
    <div>
      <div className="flex flex-wrap items-end justify-between gap-4">
        <div className="max-w-2xl">
          <p className="text-xs font-bold uppercase tracking-wide text-[#1768ad]">Ulasan Pengunjung</p>
          <h1 className="mt-2 font-heading text-3xl font-bold text-[#17324d] sm:text-4xl">
            Apa kata pengunjung Kawah Putih
          </h1>
          <p className="mt-3 text-[15px] leading-relaxed text-[#698097]">
            Rating dan ulasan hanya berasal dari pengunjung dengan booking yang sudah dibayar.
          </p>
        </div>

        {summary && summary.total_reviews > 0 && (
          <div className="flex items-center gap-4 rounded-md border border-[#d7e3ed] bg-white px-5 py-4">
            <span className="text-4xl font-bold leading-none text-[#17324d]">
              {averageRating.toFixed(1)}
            </span>
            <div>
              <RatingStars value={Math.round(averageRating)} />
              <p className="mt-1 text-xs text-[#698097]">{displayTotal} ulasan</p>
            </div>
          </div>
        )}
      </div>

      <div className="mt-6 flex flex-col gap-3 rounded-lg border border-[#d7e3ed] bg-white p-4 lg:flex-row lg:items-center lg:justify-between">
        <div className="flex flex-wrap items-center gap-2">
          <span className="mr-1 text-xs font-semibold text-[#698097]">Filter rating</span>
          <button
            type="button"
            onClick={() => setRating(null)}
            className={cn(
              "h-8 rounded-md border px-3 text-xs font-semibold transition-colors",
              rating === null
                ? "border-[#1768ad] bg-[#1768ad] text-white"
                : "border-[#d7e3ed] bg-white text-[#29445e] hover:bg-[#f2f7fb]"
            )}
          >
            Semua
          </button>
          {RATING_OPTIONS.map((value) => (
            <button
              key={value}
              type="button"
              aria-pressed={rating === value}
              onClick={() => setRating((current) => (current === value ? null : value))}
              className={cn(
                "inline-flex h-8 items-center gap-1 rounded-md border px-3 text-xs font-semibold transition-colors",
                rating === value
                  ? "border-[#1768ad] bg-[#1768ad] text-white"
                  : "border-[#d7e3ed] bg-white text-[#29445e] hover:bg-[#f2f7fb]"
              )}
            >
              {value}
              <Star
                className={cn(
                  "h-3 w-3",
                  rating === value ? "fill-white text-white" : "fill-amber-400 text-amber-400"
                )}
                aria-hidden="true"
              />
            </button>
          ))}
        </div>

        <label className="flex items-center gap-2 text-xs font-semibold text-[#698097]">
          Urutkan
          <select
            value={sort}
            onChange={(event) => setSort(event.target.value as ReviewSort)}
            className="h-8 rounded-md border border-[#d7e3ed] bg-white px-2 text-xs font-semibold text-[#29445e] outline-none focus:border-[#1768ad]"
            aria-label="Urutkan ulasan"
          >
            {SORT_OPTIONS.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </label>
      </div>

      {reviews !== null && !error && (
        <p className="mt-3 text-xs text-[#698097]">
          {total} ulasan{rating ? ` dengan rating ${rating} bintang` : ""} · diurutkan:{" "}
          {SORT_LABEL[sort]}
        </p>
      )}

      {reviews === null ? (
        <div className="mt-4 grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
          {Array.from({ length: 6 }, (_, index) => (
            <div
              key={index}
              className="h-44 animate-pulse rounded-md border border-[#d7e3ed] bg-white"
            />
          ))}
        </div>
      ) : error ? (
        <p className="mt-4 rounded-md border border-[#d7e3ed] bg-white px-5 py-4 text-sm text-[#698097]">
          Ulasan belum dapat dimuat. Silakan coba muat ulang halaman.
        </p>
      ) : reviews.length === 0 ? (
        <div className="mt-4 rounded-md border border-[#d7e3ed] bg-white px-5 py-4 text-sm text-[#698097]">
          {rating ? (
            <>
              Belum ada ulasan dengan rating <b>{rating}</b> bintang.{" "}
              <button
                type="button"
                onClick={() => setRating(null)}
                className="font-semibold text-[#1768ad] hover:underline"
              >
                Tampilkan semua ulasan
              </button>
            </>
          ) : (
            "Belum ada ulasan. Bagikan pengalamanmu dan jadilah yang pertama menulis ulasan."
          )}
        </div>
      ) : (
        <div className="mt-4 grid gap-4 sm:grid-cols-2 xl:grid-cols-3">
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
