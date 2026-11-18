// Proxy runtime berkas unggahan backend (/uploads/*) lewat origin frontend.
// Hanya GET/HEAD; berkas tidak pernah diubah dari sisi frontend.

import type { NextRequest } from "next/server";

import { proxyToBackend } from "@/lib/backend-proxy";

export const dynamic = "force-dynamic";
export const runtime = "nodejs";

type RouteContext = { params: Promise<{ path: string[] }> };

async function handle(request: NextRequest, context: RouteContext) {
  const method = request.method.toUpperCase();
  if (method !== "GET" && method !== "HEAD") {
    return Response.json(
      { success: false, message: "Metode tidak didukung." },
      { status: 405, headers: { Allow: "GET, HEAD" } }
    );
  }

  const { path } = await context.params;
  return proxyToBackend(request, "/uploads", path);
}

export { handle as GET, handle as HEAD };
