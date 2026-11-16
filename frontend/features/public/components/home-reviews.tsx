"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import Link from "next/link";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { ArrowRight, Loader2, Star } from "lucide-react";
import { toast } from "sonner";
import { bookingService } from "@/services/booking.api.service";
import { reviewService } from "@/services/review.api.service";
import { useAuthReady } from "@/providers/auth-provider";
import { useAuthStore } from "@/store/auth.store";
import { cn } from "@/lib/utils";
import type { Booking } from "@/types/api";
import type { Review, ReviewSummary } from "@/types/review";
import { ReviewCard, RatingStars } from "@/features/review/components/review-display";
import {
  COMMENT_MAX,
  COMMENT_MIN,
  RATING_MAX,
  reviewSchema,
  type ReviewFormValues,
} from "@/features/review/schemas/review.schema";

const REVIEWABLE_STATUSES = new Set(["PAID", "CONFIRMED", "USED"]);
const HOME_REVIEW_LIMIT = 3;

const FIELD_CLASS =
  "w-full rounded-md border border-[#d7e3ed] bg-white text-sm text-[#17324d] outline-none transition placeholder:text-[#94A3B8] focus:border-[#1768ad] focus:ring-1 focus:ring-[#1768ad]";

export function HomeReviews() {
  const ready = useAuthReady();
  const user = useAuthStore((s) => s.user);
  const userId = user?.id ?? null;

  const [reviews, setReviews] = useState<Review[] | null>(null);
  const [totalReviews, setTotalReviews] = useState(0);
  const [summary, setSummary] = useState<ReviewSummary | null>(null);
  const [reviewsError, setReviewsError] = useState(false);
  const [eligibleBookings, setEligibleBookings] = useState<Booking[] | null>(null);
  const [myReviewCount, setMyReviewCount] = useState(0);
  const [formError, setFormError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  const {
    register,
    handleSubmit,
    reset,
    setValue,
    watch,
    formState: { errors },
  } = useForm<ReviewFormValues>({
    resolver: zodResolver(reviewSchema),
    mode: "onTouched",
    defaultValues: { booking_id: "", rating: 0, comment: "" },
  });

  const rating = watch("rating") ?? 0;
  const commentLength = (watch("comment") ?? "").length;

  useEffect(() => {
    let alive = true;
    reviewService
      .listApproved({ limit: 100 })
      .then(({ items, total }) => {
        if (!alive) return;
        setReviews(items);
        setTotalReviews(total);
      })
      .catch(() => {
        if (!alive) return;
        setReviews([]);
        setReviewsError(true);
      });
    reviewService
      .summary()
      .then((data) => {
        if (!alive) return;
        setSummary(data);
      })
      .catch(() => {
        // Ringkasan gagal dimuat — rata-rata dihitung dari daftar ulasan yang tampil.
      });
    return () => {
      alive = false;
    };
  }, []);

  const refreshSummary = useCallback(async () => {
    try {
      setSummary(await reviewService.summary());
    } catch {
      // Biarkan rata-rata dihitung dari daftar ulasan yang sudah diperbarui.
    }
  }, []);

  const loadMine = useCallback(async () => {
    if (!userId) {
      setEligibleBookings(null);
      setMyReviewCount(0);
      setFormError(null);
      setSuccess(null);
      return;
    }
    setFormError(null);
    setSuccess(null);
    try {
      const [bookings, myReviews] = await Promise.all([
        bookingService.listMine({ limit: 100 }),
        reviewService.listMine(),
      ]);
      const reviewedBookingIds = new Set(
        myReviews.map((review) => review.booking_id).filter((id): id is string => Boolean(id))
      );
      setMyReviewCount(myReviews.length);
      setEligibleBookings(
        bookings.filter(
          (booking) => REVIEWABLE_STATUSES.has(booking.status) && !reviewedBookingIds.has(booking.id)
        )
      );
    } catch (e) {
      setEligibleBookings([]);
      setFormError(e instanceof Error ? e.message : "Gagal memuat data booking");
    }
  }, [userId]);

  useEffect(() => {
    if (!ready) return;
    void loadMine();
  }, [ready, loadMine]);

  const averageRating = useMemo(() => {
    if (summary && summary.total_reviews > 0) return summary.average_rating;
    if (!reviews || reviews.length === 0) return 0;
    return reviews.reduce((sum, review) => sum + review.rating, 0) / reviews.length;
  }, [summary, reviews]);

  const displayTotal = summary?.total_reviews ?? totalReviews;

  async function onSubmit(values: ReviewFormValues) {
    if (!user) return;
    setSubmitting(true);
    setFormError(null);
    setSuccess(null);
    try {
      const created = await reviewService.create(values);
      if (created.is_approved) {
        setReviews((items) => [created, ...(items ?? [])]);
        setTotalReviews((count) => count + 1);
        setMyReviewCount((count) => count + 1);
        void refreshSummary();
      }
      setSuccess("Terima kasih! Ulasanmu sudah tayang di halaman ini.");
      toast.success("Ulasan terkirim", {
        description: "Ulasanmu langsung tayang di halaman ini.",
      });
      setEligibleBookings((items) => (items ?? []).filter((booking) => booking.id !== values.booking_id));
      reset({ booking_id: "", rating: 0, comment: "" });
    } catch (e) {
      const message = e instanceof Error ? e.message : "Gagal mengirim ulasan";
      setFormError(message);
      toast.error("Gagal mengirim ulasan", { description: message });
    } finally {
      setSubmitting(false);
    }
  }

  const showReviewSkeleton = reviews === null;
  const formPending = !ready || (!!user && eligibleBookings === null);
  const visibleReviews = (reviews ?? []).slice(0, HOME_REVIEW_LIMIT);

  return (
    <section id="ulasan" className="border-y border-[#d7e3ed] bg-white">
      <div className="page-shell py-14">
        <div className="grid gap-8 lg:grid-cols-[1.45fr_1fr]">
          <div>
            <p className="text-xs font-bold uppercase tracking-wide text-[#1768ad]">Ulasan Pengunjung</p>
            <div className="mt-2 flex flex-wrap items-end justify-between gap-4">
              <div>
                <h2 className="font-heading text-2xl font-bold text-[#17324d]">Cerita dari pengunjung</h2>
                <p className="mt-2 max-w-xl text-[15px] leading-relaxed text-[#698097]">
                  Rating dan ulasan dari pengunjung yang sudah berkunjung ke Kawah Putih.
                </p>
              </div>

              {reviews !== null && reviews.length > 0 && (
                <div className="flex items-center gap-3 rounded-md border border-[#d7e3ed] bg-[#f7fbff] px-4 py-3">
                  <span className="text-3xl font-bold leading-none text-[#17324d]">
                    {averageRating.toFixed(1)}
                  </span>
                  <div>
                    <RatingStars value={Math.round(averageRating)} starClassName="h-3.5 w-3.5" />
                    <p className="mt-1 text-xs text-[#698097]">{displayTotal} ulasan</p>
                  </div>
                </div>
              )}
            </div>

            {showReviewSkeleton ? (
              <div className="mt-6 grid gap-3 sm:grid-cols-2">
                {Array.from({ length: HOME_REVIEW_LIMIT }, (_, i) => (
                  <div key={i} className="h-36 animate-pulse rounded-md border border-[#d7e3ed] bg-[#f7fbff]" />
                ))}
              </div>
            ) : reviewsError ? (
              <p className="mt-6 rounded-md border border-[#d7e3ed] bg-[#f7fbff] px-4 py-3 text-sm text-[#698097]">
                Ulasan belum dapat dimuat. Silakan coba muat ulang halaman.
              </p>
            ) : visibleReviews.length === 0 ? (
              <p className="mt-6 rounded-md border border-[#d7e3ed] bg-[#f7fbff] px-4 py-3 text-sm text-[#698097]">
                Belum ada ulasan. Bagikan pengalamanmu dan jadilah yang pertama menulis ulasan.
              </p>
            ) : (
              <div className="mt-6 grid gap-3 sm:grid-cols-2">
                {visibleReviews.map((review) => (
                  <ReviewCard key={review.id} review={review} clamp />
                ))}
              </div>
            )}

            {reviews !== null && reviews.length > 0 && (
              <Link
                href="/review"
                className="mt-5 inline-flex items-center gap-2 text-sm font-semibold text-[#1768ad] hover:underline"
              >
                Lihat semua ulasan <ArrowRight className="h-4 w-4" aria-hidden="true" />
              </Link>
            )}
          </div>

          <div className="h-fit rounded-lg border border-[#1768ad]/20 bg-[#f7fbff] p-5">
            <h3 className="text-base font-bold text-[#17324d]">Bagikan pengalamanmu</h3>
            <p className="mt-1 text-xs leading-relaxed text-[#698097]">
              Rating 1-5 dan ulasan {COMMENT_MIN}-{COMMENT_MAX} karakter. Ulasan langsung tayang di
              halaman ini.
            </p>

            {success && (
              <p className="mt-4 rounded-md border border-[#ccebd9] bg-[#edf9f2] px-3 py-2 text-xs leading-relaxed text-[#16876a]">
                {success}
              </p>
            )}

            {formPending ? (
              <div className="mt-5 space-y-3" aria-label="Memuat formulir ulasan">
                <div className="h-10 animate-pulse rounded-md bg-[#e7f2fb]" />
                <div className="h-10 animate-pulse rounded-md bg-[#e7f2fb]" />
                <div className="h-24 animate-pulse rounded-md bg-[#e7f2fb]" />
              </div>
            ) : !user ? (
              <div className="mt-5">
                <p className="text-sm leading-relaxed text-[#29445e]">
                  Masuk terlebih dahulu untuk memberi rating dan ulasan. Ulasan hanya bisa ditulis untuk
                  booking milikmu yang sudah dibayar.
                </p>
                <Link
                  href={`/login?redirect=${encodeURIComponent("/#ulasan")}`}
                  className="mt-4 inline-flex h-10 items-center justify-center rounded-md bg-[#1768ad] px-5 text-sm font-semibold text-white hover:bg-[#14548f]"
                >
                  Masuk untuk Menulis Ulasan
                </Link>
                <p className="mt-3 text-xs text-[#698097]">
                  Belum punya akun?{" "}
                  <Link href="/register" className="font-semibold text-[#1768ad] hover:underline">
                    Daftar sekarang
                  </Link>
                </p>
              </div>
            ) : eligibleBookings && eligibleBookings.length === 0 ? (
              <div className="mt-5">
                {formError && (
                  <p className="mb-3 rounded-md bg-[#fdeaea] px-3 py-2 text-xs leading-relaxed text-[#c53030]">
                    {formError}
                  </p>
                )}
                <p className="text-sm leading-relaxed text-[#29445e]">
                  {myReviewCount > 0
                    ? "Semua booking yang bisa diulas sudah kamu nilai. Terima kasih sudah berbagi pengalaman!"
                    : "Belum ada booking yang bisa diulas. Ulasan tersedia setelah pembayaran dikonfirmasi dan booking belum pernah dinilai."}
                </p>
                <div className="mt-4">
                  <Link
                    href="/review"
                    className="inline-flex h-9 items-center rounded-md border border-[#d7e3ed] bg-white px-4 text-xs font-semibold text-[#29445e] hover:bg-[#f2f7fb]"
                  >
                    Lihat Semua Ulasan
                  </Link>
                </div>
              </div>
            ) : (
              <form onSubmit={handleSubmit(onSubmit)} className="mt-5 space-y-4" noValidate>
                {formError && (
                  <p className="rounded-md bg-[#fdeaea] px-3 py-2 text-xs leading-relaxed text-[#c53030]">
                    {formError}
                  </p>
                )}

                <div>
                  <label htmlFor="home-review-booking" className="mb-1 block text-xs font-semibold text-[#29445e]">
                    Booking
                  </label>
                  <select
                    id="home-review-booking"
                    className={cn(FIELD_CLASS, "h-10 px-3")}
                    aria-invalid={!!errors.booking_id}
                    aria-describedby={errors.booking_id ? "home-review-booking-error" : undefined}
                    {...register("booking_id")}
                  >
                    <option value="">Pilih booking</option>
                    {(eligibleBookings ?? []).map((booking) => (
                      <option key={booking.id} value={booking.id}>
                        {booking.booking_code} - {booking.visit_date}
                      </option>
                    ))}
                  </select>
                  {errors.booking_id && (
                    <p id="home-review-booking-error" className="mt-1 text-xs text-[#c53030]">
                      {errors.booking_id.message}
                    </p>
                  )}
                </div>

                <div>
                  <span className="mb-1 block text-xs font-semibold text-[#29445e]">Rating</span>
                  <div className="flex flex-wrap items-center gap-1.5" role="group" aria-label="Pilih rating">
                    {Array.from({ length: RATING_MAX }, (_, i) => {
                      const value = i + 1;
                      return (
                        <button
                          key={value}
                          type="button"
                          aria-label={`${value} bintang`}
                          aria-pressed={rating === value}
                          onClick={() => setValue("rating", value, { shouldValidate: true, shouldTouch: true })}
                          className="rounded p-0.5 transition-transform hover:scale-110 focus:outline-none focus:ring-2 focus:ring-[#1768ad]"
                        >
                          <Star
                            className={cn(
                              "h-7 w-7",
                              value <= rating ? "fill-amber-400 text-amber-400" : "text-[#c8dded]"
                            )}
                          />
                        </button>
                      );
                    })}
                    <span className="ml-1 text-xs text-[#698097]">
                      {rating > 0 ? `${rating}/${RATING_MAX}` : "Belum dipilih"}
                    </span>
                  </div>
                  {errors.rating && <p className="mt-1 text-xs text-[#c53030]">{errors.rating.message}</p>}
                </div>

                <div>
                  <label htmlFor="home-review-comment" className="mb-1 block text-xs font-semibold text-[#29445e]">
                    Ulasan
                  </label>
                  <textarea
                    id="home-review-comment"
                    rows={4}
                    maxLength={COMMENT_MAX}
                    placeholder="Ceritakan pengalaman kunjunganmu..."
                    className={cn(FIELD_CLASS, "px-3 py-2")}
                    aria-invalid={!!errors.comment}
                    aria-describedby={errors.comment ? "home-review-comment-error" : undefined}
                    {...register("comment")}
                  />
                  <div className="mt-1 flex items-start justify-between gap-2">
                    {errors.comment ? (
                      <p id="home-review-comment-error" className="text-xs text-[#c53030]">
                        {errors.comment.message}
                      </p>
                    ) : (
                      <span className="text-xs text-[#698097]">Minimal {COMMENT_MIN} karakter</span>
                    )}
                    <span className="shrink-0 text-xs tabular-nums text-[#94A3B8]">
                      {commentLength}/{COMMENT_MAX}
                    </span>
                  </div>
                </div>

                <button
                  type="submit"
                  disabled={submitting}
                  className="inline-flex h-10 w-full items-center justify-center gap-2 rounded-md bg-[#1768ad] px-5 text-sm font-semibold text-white hover:bg-[#14548f] disabled:pointer-events-none disabled:opacity-60"
                >
                  {submitting && <Loader2 className="h-4 w-4 animate-spin" aria-hidden="true" />}
                  {submitting ? "Mengirim..." : "Kirim Ulasan"}
                </button>
              </form>
            )}
          </div>
        </div>
      </div>
    </section>
  );
}
