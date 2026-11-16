"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { Star, Trash2 } from "lucide-react";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { Select } from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { PageHeader, SectionCard, StatusPill } from "@/components/panel/panel-ui";
import { bookingService } from "@/services/booking.api.service";
import { reviewService } from "@/services/review.api.service";
import { cn } from "@/lib/utils";
import type { Booking } from "@/types/api";
import { REVIEW_STATUS_LABEL, type Review } from "@/types/review";
import { RatingStars, formatReviewDate } from "./review-display";
import {
  COMMENT_MAX,
  RATING_MAX,
  reviewSchema,
  type ReviewFormValues,
} from "../schemas/review.schema";

const REVIEWABLE_STATUSES = new Set(["PAID", "CONFIRMED", "USED"]);

export function ReviewManager() {
  const [bookings, setBookings] = useState<Booking[] | null>(null);
  const [reviews, setReviews] = useState<Review[] | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [deletingId, setDeletingId] = useState<string | null>(null);

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

  const rating = watch("rating");

  const load = useCallback(async () => {
    setError(null);
    try {
      const [myBookings, myReviews] = await Promise.all([
        bookingService.listMine({ limit: 100 }),
        reviewService.listMine(),
      ]);
      setBookings(myBookings);
      setReviews(myReviews);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Gagal memuat data ulasan");
    }
  }, []);

  useEffect(() => {
    void load();
  }, [load]);

  const reviewedBookingIds = useMemo(
    () => new Set((reviews ?? []).map((review) => review.booking_id).filter(Boolean)),
    [reviews]
  );

  const reviewableBookings = useMemo(
    () =>
      (bookings ?? [])
        .filter((booking) => REVIEWABLE_STATUSES.has(booking.status))
        .filter((booking) => !reviewedBookingIds.has(booking.id)),
    [bookings, reviewedBookingIds]
  );

  const bookingCodes = useMemo(() => {
    const map = new Map<string, string>();
    for (const booking of bookings ?? []) map.set(booking.id, booking.booking_code);
    return map;
  }, [bookings]);

  async function onSubmit(values: ReviewFormValues) {
    setSubmitting(true);
    setError(null);
    try {
      await reviewService.create(values);
      toast.success("Ulasan terkirim dan langsung tayang");
      reset({ booking_id: "", rating: 0, comment: "" });
      const myReviews = await reviewService.listMine();
      setReviews(myReviews);
    } catch (e) {
      const message = e instanceof Error ? e.message : "Gagal mengirim ulasan";
      setError(message);
      toast.error(message);
    } finally {
      setSubmitting(false);
    }
  }

  async function handleDelete(review: Review) {
    if (!window.confirm("Hapus ulasan ini?")) return;
    setDeletingId(review.id);
    try {
      await reviewService.remove(review.id);
      setReviews((items) => (items ?? []).filter((item) => item.id !== review.id));
      toast.success("Ulasan dihapus");
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "Gagal menghapus ulasan");
    } finally {
      setDeletingId(null);
    }
  }

  const loading = bookings === null || reviews === null;

  return (
    <div>
      <PageHeader title="Ulasan Saya" subtitle="Beri rating dan ulasan untuk kunjungan yang sudah kamu lakukan" />

      {error && (
        <p className="mb-3 rounded-md bg-[#fdeaea] px-3 py-2 text-sm text-[#c53030]">{error}</p>
      )}

      <div className="grid gap-4 lg:grid-cols-[minmax(0,420px)_1fr]">
        <SectionCard title="Tulis Ulasan" subtitle="Rating 1-5 dan ulasan 10-1000 karakter">
          {loading ? (
            <div className="h-40 animate-pulse rounded-md bg-[#eef5fb]" />
          ) : reviewableBookings.length === 0 ? (
            <p className="text-sm text-[#698097]">
              Belum ada booking yang bisa diulas. Ulasan tersedia setelah pembayaran dikonfirmasi dan belum pernah dinilai.
            </p>
          ) : (
            <form onSubmit={handleSubmit(onSubmit)} className="space-y-4" noValidate>
              <div>
                <label htmlFor="review-booking" className="mb-1 block text-xs font-semibold text-[#29445e]">
                  Booking
                </label>
                <Select id="review-booking" {...register("booking_id")}>
                  <option value="">Pilih booking</option>
                  {reviewableBookings.map((booking) => (
                    <option key={booking.id} value={booking.id}>
                      {booking.booking_code} - {booking.visit_date}
                    </option>
                  ))}
                </Select>
                {errors.booking_id && <p className="mt-1 text-xs text-[#c53030]">{errors.booking_id.message}</p>}
              </div>

              <div>
                <span className="mb-1 block text-xs font-semibold text-[#29445e]">Rating</span>
                <div className="flex items-center gap-1.5" role="group" aria-label="Pilih rating">
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
                            "h-8 w-8",
                            value <= rating ? "fill-amber-400 text-amber-400" : "text-[#c8dded]"
                          )}
                        />
                      </button>
                    );
                  })}
                  <span className="ml-1 text-xs text-[#698097]">{rating > 0 ? `${rating}/${RATING_MAX}` : "Belum dipilih"}</span>
                </div>
                {errors.rating && <p className="mt-1 text-xs text-[#c53030]">{errors.rating.message}</p>}
              </div>

              <div>
                <label htmlFor="review-comment" className="mb-1 block text-xs font-semibold text-[#29445e]">
                  Ulasan
                </label>
                <Textarea
                  id="review-comment"
                  rows={4}
                  maxLength={COMMENT_MAX}
                  placeholder="Ceritakan pengalaman kunjunganmu..."
                  {...register("comment")}
                />
                <div className="mt-1 flex items-start justify-between gap-2">
                  {errors.comment ? (
                    <p className="text-xs text-[#c53030]">{errors.comment.message}</p>
                  ) : (
                    <span className="text-xs text-[#698097]">Minimal 10 karakter</span>
                  )}
                  <span className="shrink-0 text-xs tabular-nums text-[#94A3B8]">
                    {watch("comment").length}/{COMMENT_MAX}
                  </span>
                </div>
              </div>

              <Button type="submit" loading={submitting} className="w-full">
                Kirim Ulasan
              </Button>
            </form>
          )}
        </SectionCard>

        <SectionCard title="Riwayat Ulasan" subtitle="Ulasan yang sudah kamu kirim">
          {loading ? (
            <div className="h-40 animate-pulse rounded-md bg-[#eef5fb]" />
          ) : (reviews ?? []).length === 0 ? (
            <p className="text-sm text-[#698097]">Kamu belum menulis ulasan.</p>
          ) : (
            <ul className="divide-y divide-[#eef2f6]">
              {(reviews ?? []).map((review) => (
                <li key={review.id} className="py-3 first:pt-0 last:pb-0">
                  <div className="flex flex-wrap items-start justify-between gap-2">
                    <div>
                      <p className="text-xs font-semibold text-[#17324d]">
                        {review.booking_id ? bookingCodes.get(review.booking_id) ?? "Booking" : "Ulasan"}
                      </p>
                      <p className="mt-0.5 text-[10px] text-[#698097]">{formatReviewDate(review.created_at)}</p>
                    </div>
                    <div className="flex items-center gap-2">
                      <StatusPill status={REVIEW_STATUS_LABEL[review.status]} />
                      <button
                        type="button"
                        onClick={() => handleDelete(review)}
                        disabled={deletingId === review.id}
                        className="rounded p-1 text-[#c53030] hover:bg-[#fdeaea] disabled:opacity-50"
                        aria-label="Hapus ulasan"
                      >
                        <Trash2 className="h-4 w-4" />
                      </button>
                    </div>
                  </div>
                  <RatingStars value={review.rating} className="mt-2" />
                  <p className="mt-1 text-sm leading-relaxed text-[#29445e]">{review.comment}</p>
                </li>
              ))}
            </ul>
          )}
        </SectionCard>
      </div>
    </div>
  );
}
