# syntax=docker/dockerfile:1

# Build tools run natively; the Go binary targets the requested image platform.
FROM --platform=$BUILDPLATFORM golang:1.27.1-alpine AS builder-go

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY backend/ ./backend/

ARG TARGETOS
ARG TARGETARCH
RUN test -n "$TARGETOS" && test -n "$TARGETARCH" \
    && CGO_ENABLED=0 GOOS="$TARGETOS" GOARCH="$TARGETARCH" \
       go build -trimpath -ldflags="-s -w" -o /out/tas-server ./backend

# Build the Node.js application
FROM --platform=$BUILDPLATFORM node:26.9-alpine AS builder-node

WORKDIR /app

COPY frontend/package*.json ./
RUN npm ci

COPY frontend/ ./
RUN npm run build

# Final image
FROM alpine:latest

RUN apk --no-cache add ca-certificates wget \
    && addgroup -S -g 10001 tas \
    && adduser -S -D -H -u 10001 -G tas tas

WORKDIR /app

COPY --from=builder-go /out/tas-server /app/tas-server

COPY --from=builder-node /app/dist ./frontend/dist
COPY config.example.toml ./config.example.toml

USER tas

# Execute the actual binary in each target image before it can be published.
# This checks both executable format and architecture, without config or a DB.
ARG TARGETOS
ARG TARGETARCH
RUN test "$(/app/tas-server --build-info)" = "$TARGETOS/$TARGETARCH" \
    && test -s /app/frontend/dist/index.html

EXPOSE 2005

HEALTHCHECK --interval=30s --timeout=5s --start-period=30s --retries=3 \
  CMD wget --quiet --tries=1 --timeout=4 --output-document=/dev/null http://127.0.0.1:2005/healthcheck || exit 1

ENTRYPOINT ["/app/tas-server"]
