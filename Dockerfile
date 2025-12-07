FROM --platform=$BUILDPLATFORM oven/bun:1.3-alpine AS frontend-builder
WORKDIR /app
COPY frontend/package.json frontend/bun.lock* ./
RUN bun install --frozen-lockfile
COPY frontend/ .
RUN bun run build

FROM golang:1.24-alpine AS backend-builder
WORKDIR /app
RUN apk add --no-cache gcc musl-dev
COPY backend/go.mod backend/go.sum* ./
RUN go mod download
COPY backend/ .
RUN CGO_ENABLED=1 CGO_CFLAGS="-D_LARGEFILE64_SOURCE" go build -ldflags="-s -w" -tags musl -o caddy-proxy-manager main.go

FROM caddy:builder AS caddy-builder
RUN xcaddy build \
    --with github.com/caddy-dns/cloudflare \
    --with github.com/mholt/caddy-l4

FROM caddy:2.10.2-alpine
WORKDIR /app
COPY --from=caddy-builder /usr/bin/caddy /usr/bin/caddy
COPY --from=backend-builder /app/caddy-proxy-manager /app/caddy-proxy-manager
COPY --from=frontend-builder /app/dist /app/public
COPY Caddyfile /etc/caddy/Caddyfile
COPY default_page/ /app/default/
COPY start.sh /app/start.sh
RUN mkdir -p /data && chmod +x /app/start.sh && apk add --no-cache curl
EXPOSE 80 81 443 2019
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
    CMD curl -f http://localhost:8080/api/health || exit 1
CMD ["/app/start.sh"]

