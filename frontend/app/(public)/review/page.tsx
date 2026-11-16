import type { Metadata } from "next";
import { PublicReviewList } from "@/features/public/components/public-review-list";

export const metadata: Metadata = {
  title: "Ulasan",
  description: "Ulasan pengunjung Kawah Putih Rancabali.",
};

export default function PublicReviewPage() {
  return (
    <section className="page-shell py-12">
      <PublicReviewList />
    </section>
  );
}
