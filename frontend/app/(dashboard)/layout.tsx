"use client";

import { useRouteGuard } from "@/hooks/use-route-guard";
import { PanelLayout, type PanelLink } from "@/components/layout/panel-layout";
import { Banknote, CalendarClock, History, LayoutDashboard, Settings, Star, Ticket, TicketCheck, User, Wallet } from "lucide-react";

// Sidebar pengunjung — memakai kerangka panel yang sama dengan role lain.
const LINKS: PanelLink[] = [
  { href: "/dashboard", label: "Ringkasan", icon: LayoutDashboard },
  { href: "/dashboard/booking", label: "Booking Saya", icon: Ticket },
  { href: "/dashboard/ticket", label: "Tiket Saya", icon: TicketCheck },
  { href: "/dashboard/payment", label: "Pembayaran", icon: Wallet },
  { href: "/dashboard/refund", label: "Refund", icon: Banknote },
  { href: "/dashboard/reschedule", label: "Reschedule", icon: CalendarClock },
  { href: "/dashboard/history", label: "Riwayat", icon: History },
  { href: "/dashboard/review", label: "Ulasan Saya", icon: Star },
  { href: "/dashboard/profile", label: "Profil", icon: User },
  { href: "/dashboard/setting", label: "Keamanan Akun", icon: Settings },
];

export default function DashboardLayout({ children }: { children: React.ReactNode }) {
  const guard = useRouteGuard(["PENGUNJUNG", "ADMIN", "MANAGER", "PETUGAS"]);

  if (guard !== "allowed") return null;

  return (
    <PanelLayout role="Pengunjung" links={LINKS}>
      {children}
    </PanelLayout>
  );
}
