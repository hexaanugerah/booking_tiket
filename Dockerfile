# syntax=docker/dockerfile:1
#
# Dockerfile root Kawah Putih: berisi dua target image.
#   - backend  : binary Go (API server) + runtime Alpine
#   - frontend : Next.js yang sudah di-build + runtime Node.js
#
# Build manual:
#   docker build --target backend  -t kawahputih-backend  .
#   docker build --target frontend -t kawahputih-frontend .
#
# Docker Compose memilih target yang tepat secara otomatis
# (lihat docker-compose.yml di folder ini).

# ============================================================================
# BACKEND
# ============================================================================
FROM golang:1.23-alpine AS backend-builder

RUN apk add --no-cache git ca-certificates
WORKDIR /src

# Cache modul Go terpisah dari source supaya rebuild lebih cepat.
COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/server \
    ./cmd/server

FROM alpine:3.22 AS backend

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S app \
    && adduser -S app -G app

WORKDIR /app
COPY --from=backend-builder /out/server ./server

# Folder unggahan harus milik user non-root. Saat volume bernama dipasang,
# kepemilikan ini ikut disalin oleh Docker.
RUN mkdir -p /app/uploads && chown -R app:app /app

USER app
EXPOSE 8080

ENTRYPOINT ["/app/server"]

# ============================================================================
# FRONTEND
# ============================================================================
FROM node:22-alpine AS frontend-deps

WORKDIR /app
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci

FROM node:22-alpine AS frontend-builder

WORKDIR /app
ENV NEXT_TELEMETRY_DISABLED=1

COPY --from=frontend-deps /app/node_modules ./node_modules
COPY frontend/ ./

# NEXT_PUBLIC_* di-inline saat build. Nilainya harus berupa URL absolut yang
# bisa dijangkau browser (bukan nama service Docker seperti "backend").
# Ubah lewat build arg di docker-compose.yml / .env.
ARG NEXT_PUBLIC_API_BASE_URL=http://localhost:8080/api/v1
ENV NEXT_PUBLIC_API_BASE_URL=$NEXT_PUBLIC_API_BASE_URL

# Cache build webpack tidak diperlukan saat runtime dan besarnya bisa
# ratusan MB, jadi dibuang sebelum disalin ke image akhir.
RUN npm run build && rm -rf .next/cache

FROM node:22-alpine AS frontend-prod-deps

WORKDIR /app
COPY frontend/package.json frontend/package-lock.json ./
# typescript dibutuhkan untuk memuat next.config.ts saat `next start`.
# Package ini ada di devDependencies sehingga tidak ikut `npm ci --omit=dev`,
# jadi disalin langsung dari node_modules hasil build.
RUN npm ci --omit=dev

COPY --from=frontend-builder /app/node_modules/typescript ./node_modules/typescript

FROM node:22-alpine AS frontend

WORKDIR /app
ENV NODE_ENV=production \
    NEXT_TELEMETRY_DISABLED=1 \
    PORT=3000

# --chown dipakai langsung di COPY supaya tidak perlu `chown -R` yang akan
# menyalin ulang seluruh isi /app ke layer baru (membuat image membengkak).
COPY --chown=node:node --from=frontend-prod-deps /app/node_modules ./node_modules
COPY --chown=node:node --from=frontend-builder /app/.next ./.next
COPY --chown=node:node --from=frontend-builder /app/public ./public
COPY --chown=node:node frontend/package.json frontend/package-lock.json frontend/next.config.ts frontend/tsconfig.json ./

USER node

EXPOSE 3000
CMD ["npm", "run", "start"]
