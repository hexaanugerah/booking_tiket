// Proxy runtime dari route handler Next.js ke backend Go.
//
// Dipakai oleh app/api/v1/[...path]/route.ts dan app/uploads/[...path]/route.ts.
// Berbeda dari rewrites di next.config.ts yang dibekukan saat build, target
// proxy di sini dibaca dari process.env pada setiap request sehingga URL
// backend bisa diubah lewat variabel Worker tanpa build ulang.

import type { NextRequest } from "next/server";

import {
  BackendOriginError,
  resolveBackendOrigin,
} from "@/lib/backend-origin";

// Header hop-by-hop tidak boleh diteruskan (RFC 7230).
const HOP_BY_HOP_HEADERS = new Set([
  "connection",
  "keep-alive",
  "proxy-authenticate",
  "proxy-authorization",
  "te",
  "trailer",
  "transfer-encoding",
  "upgrade",
]);

// Header yang dihitung ulang oleh proxy, bukan disalin dari browser.
const REQUEST_HEADERS_TO_REPLACE = new Set([
  "host",
  "content-length",
  "accept-encoding",
  "x-forwarded-for",
  "x-forwarded-host",
  "x-forwarded-proto",
  "x-real-ip",
  "forwarded",
]);

function configErrorResponse(error: BackendOriginError): Response {
  return Response.json(
    {
      success: false,
      message: error.message,
      errors: { code: "BACKEND_ORIGIN_UNSET" },
    },
    { status: 503 }
  );
}

function gatewayErrorResponse(error: unknown): Response {
  const detail =
    process.env.NODE_ENV !== "production" && error instanceof Error
      ? ` (${error.message})`
      : "";
  return Response.json(
    {
      success: false,
      message: `Tidak dapat terhubung ke server backend${detail}.`,
      errors: { code: "BACKEND_UNREACHABLE" },
    },
    { status: 502 }
  );
}

function forwardRequestHeaders(request: NextRequest): Headers {
  const headers = new Headers();

  request.headers.forEach((value, key) => {
    const lower = key.toLowerCase();
    if (
      HOP_BY_HOP_HEADERS.has(lower) ||
      REQUEST_HEADERS_TO_REPLACE.has(lower) ||
      lower.startsWith("cf-")
    ) {
      return;
    }
    headers.append(key, value);
  });

  // Cloudflare menulis cf-connecting-ip dari koneksi asli. Teruskan sebagai
  // X-Forwarded-For supaya rate limit backend tetap per pengunjung, bukan per
  // IP Worker. Gin v1.10 mempercayai header ini secara default.
  const clientIP =
    request.headers.get("cf-connecting-ip") ??
    request.headers.get("x-real-ip") ??
    request.headers.get("x-forwarded-for")?.split(",")[0]?.trim();
  if (clientIP) {
    headers.set("x-forwarded-for", clientIP);
    headers.set("x-real-ip", clientIP);
  }

  const incoming = new URL(request.url);
  headers.set("x-forwarded-proto", incoming.protocol.replace(":", ""));
  headers.set("x-forwarded-host", incoming.host);
  headers.set(
    "x-request-id",
    request.headers.get("x-request-id") ?? crypto.randomUUID()
  );

  return headers;
}

function browserResponse(upstream: Response): Response {
  const headers = new Headers();

  upstream.headers.forEach((value, key) => {
    const lower = key.toLowerCase();
    if (
      HOP_BY_HOP_HEADERS.has(lower) ||
      lower === "content-length" ||
      // Body dari fetch sudah didekompresi, jadi header ini tidak berlaku lagi.
      lower === "content-encoding" ||
      // Respons selalu same-origin lewat Worker; CORS backend tidak relevan.
      lower.startsWith("access-control-")
    ) {
      return;
    }
    headers.append(key, value);
  });

  // Response dengan status ini tidak boleh memiliki body.
  const hasNullBody =
    upstream.status === 204 || upstream.status === 205 || upstream.status === 304;

  return new Response(hasNullBody ? null : upstream.body, {
    status: upstream.status,
    statusText: upstream.statusText,
    headers,
  });
}

/**
 * Meneruskan request ke `${backendOrigin}${pathPrefix}/${segments...}` dengan
 * query string dan body apa adanya.
 */
export async function proxyToBackend(
  request: NextRequest,
  pathPrefix: "/api/v1" | "/uploads",
  segments: string[]
): Promise<Response> {
  let backendOrigin: string;
  try {
    backendOrigin = resolveBackendOrigin();
  } catch (error) {
    if (error instanceof BackendOriginError) return configErrorResponse(error);
    throw error;
  }

  const incoming = new URL(request.url);
  const suffix = segments.map((segment) => encodeURIComponent(segment)).join("/");
  const target = new URL(`${pathPrefix}/${suffix}`, backendOrigin);
  target.search = incoming.search;

  const method = request.method.toUpperCase();
  const hasBody = method !== "GET" && method !== "HEAD";
  const body = hasBody ? await request.arrayBuffer() : undefined;

  let upstream: Response;
  try {
    upstream = await fetch(target, {
      method,
      headers: forwardRequestHeaders(request),
      body,
      redirect: "manual",
      cache: "no-store",
    });
  } catch (error) {
    return gatewayErrorResponse(error);
  }

  return browserResponse(upstream);
}
