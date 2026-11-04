# Stage 1: Build binary
FROM golang:1.23-alpine AS builder

WORKDIR /src

RUN apk add --no-cache ca-certificates tzdata

COPY go.mod ./
COPY cmd/ ./cmd/
COPY internal/ ./internal/

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /bin/cdn-server ./cmd/server

# Stage 2: Minimal runtime image
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S cdn && adduser -S cdn -G cdn \
    && mkdir -p /data \
    && chown -R cdn:cdn /data

WORKDIR /app
COPY --from=builder /bin/cdn-server /app/cdn-server

USER cdn

ENV CDN_HOST=0.0.0.0 \
    CDN_PORT=8080 \
    CDN_STORAGE_DIR=/data \
    CDN_BASE_URL=http://localhost:8080

EXPOSE 8080
VOLUME ["/data"]

ENTRYPOINT ["/app/cdn-server"]
