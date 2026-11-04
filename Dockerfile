# Stage 1: Build binary with multi-architecture support
FROM --platform=$BUILDPLATFORM golang:1.23-alpine AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /src

RUN apk add --no-cache ca-certificates tzdata

COPY go.mod ./
COPY cmd/ ./cmd/
COPY internal/ ./internal/

RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} \
    go build -trimpath -ldflags="-s -w" -o /bin/cdn-server ./cmd/server

# Stage 2: Production runtime image optimized for Dokploy / Docker
FROM alpine:3.20

LABEL org.opencontainers.image.title="Simple Static CDN" \
      org.opencontainers.image.description="Self-hosted high-performance static file CDN and delivery engine" \
      org.opencontainers.image.source="https://github.com/Antigravity/simple-cdn" \
      org.opencontainers.image.licenses="MIT"

RUN apk add --no-cache ca-certificates tzdata wget \
    && addgroup -g 1000 -S cdn \
    && adduser -u 1000 -S cdn -G cdn \
    && mkdir -p /data \
    && chown -R cdn:cdn /data

WORKDIR /app
COPY --from=builder /bin/cdn-server /app/cdn-server

USER cdn

ENV PORT=8080 \
    CDN_PORT=8080 \
    CDN_HOST=0.0.0.0 \
    CDN_STORAGE_DIR=/data \
    CDN_BASE_URL=http://localhost:8080

EXPOSE 8080
VOLUME ["/data"]

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://127.0.0.1:${PORT:-8080}/healthz || exit 1

ENTRYPOINT ["/app/cdn-server"]
