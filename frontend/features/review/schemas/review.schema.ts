import { z } from "zod";

// Aturan ini harus sama dengan validator backend (rating 1-5, ulasan 10-1000 karakter).
export const RATING_MIN = 1;
export const RATING_MAX = 5;
export const COMMENT_MIN = 10;
export const COMMENT_MAX = 1000;

export const reviewSchema = z.object({
  booking_id: z.string().min(1, "Pilih booking yang ingin diulas").uuid("Booking tidak valid"),
  rating: z
    .number({ invalid_type_error: "Rating wajib dipilih" })
    .int("Rating harus berupa angka bulat")
    .min(RATING_MIN, "Rating wajib dipilih")
    .max(RATING_MAX, `Rating maksimal ${RATING_MAX}`),
  comment: z
    .string()
    .trim()
    .min(COMMENT_MIN, `Ulasan minimal ${COMMENT_MIN} karakter`)
    .max(COMMENT_MAX, `Ulasan maksimal ${COMMENT_MAX} karakter`),
});

export type ReviewFormValues = z.infer<typeof reviewSchema>;
