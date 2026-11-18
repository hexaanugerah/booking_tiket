// Proxy runtime seluruh endpoint backend /api/v1/* lewat origin frontend.
//
// Browser cukup memanggil origin yang sama (tidak ada CORS/mixed content),
// sedangkan Worker meneruskan request ke BACKEND_ORIGIN yang dibaca saat
// request. Jadi URL backend bisa diganti tanpa build ulang.

import type { NextRequest } from "next/server";

import { proxyToBackend } from "@/lib/backend-proxy";

export const dynamic = "force-dynamic";
export const runtime = "nodejs";

type RouteContext = { params: Promise<{ path: string[] }> };

async function handle(request: NextRequest, context: RouteContext) {
  const { path } = await context.params;
  return proxyToBackend(request, "/api/v1", path);
}

export {
  handle as GET,
  handle as POST,
  handle as PUT,
  handle as PATCH,
  handle as DELETE,
  handle as HEAD,
  handle as OPTIONS,
};
