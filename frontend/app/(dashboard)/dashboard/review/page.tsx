import type { Metadata } from "next";
import { ReviewManager } from "@/features/review/components/review-manager";

export const metadata: Metadata = { title: "Ulasan Saya" };

export default function ReviewPage() {
  return <ReviewManager />;
}
