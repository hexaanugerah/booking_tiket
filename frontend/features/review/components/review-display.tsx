import { BadgeCheck, Star, UserRound } from "lucide-react";
import { cn } from "@/lib/utils";
import type { Review } from "@/types/review";
import { RATING_MAX } from "../schemas/review.schema";

export function RatingStars({
  value,
  className,
  starClassName,
}: {
  value: number;
  className?: string;
  starClassName?: string;
}) {
  return (
    <span
      className={cn("inline-flex items-center gap-0.5", className)}
      aria-label={`${value} dari ${RATING_MAX} bintang`}
    >
      {Array.from({ length: RATING_MAX }, (_, i) => (
        <Star
          key={i}
          className={cn(
            starClassName ?? "h-4 w-4",
            i < value ? "fill-amber-400 text-amber-400" : "text-[#c8dded]"
          )}
        />
      ))}
    </span>
  );
}

export function formatReviewDate(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return date.toLocaleDateString("id-ID", { day: "2-digit", month: "short", year: "numeric" });
}

interface ReviewCardProps {
  review: Review;
  className?: string;
  clamp?: boolean;
}

export function ReviewCard({ review, className, clamp = false }: ReviewCardProps) {
  return (
    <article className={cn("rounded-md border border-[#d7e3ed] bg-[#f9fcff] p-4", className)}>
      <div className="flex items-center gap-2.5">
        <span className="grid h-8 w-8 shrink-0 place-items-center rounded-full bg-[#e7f2fb] text-[#1768ad]">
          <UserRound className="h-4 w-4" aria-hidden="true" />
        </span>
        <div className="min-w-0">
          <p className="truncate text-xs font-bold text-[#17324d]">Pengunjung Kawah Putih</p>
          <p className="inline-flex items-center gap-1 text-[10px] font-semibold text-[#16876a]">
            <BadgeCheck className="h-3 w-3" aria-hidden="true" /> Booking terverifikasi
          </p>
        </div>
      </div>

      <div className="mt-3 flex items-center justify-between gap-3">
        <RatingStars value={review.rating} starClassName="h-3.5 w-3.5" />
        <span className="shrink-0 text-[10px] text-[#94A3B8]">{formatReviewDate(review.created_at)}</span>
      </div>

      <p className={cn("mt-2 text-sm leading-relaxed text-[#29445e]", clamp && "line-clamp-4")}>
        {review.comment}
      </p>
    </article>
  );
}
