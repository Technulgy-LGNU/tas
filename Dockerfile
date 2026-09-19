# syntax=docker/dockerfile:1.7

FROM --platform=$BUILDPLATFORM node:26-bookworm AS frontend-build
WORKDIR /src/frontend

COPY frontend/package*.json ./
RUN npm ci

COPY frontend/ ./
RUN npm run build

FROM --platform=$BUILDPLATFORM golang:1.27.1-bookworm AS backend-build
WORKDIR /src
ARG TARGETOS
ARG TARGETARCH

COPY go.mod go.sum ./
RUN go mod download

COPY backend/ ./backend/
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -ldflags="-s -w" -o /out/tas-server ./backend

FROM debian:bookworm-slim AS runtime
RUN apt-get update \
	&& apt-get install -y --no-install-recommends ca-certificates \
	&& rm -rf /var/lib/apt/lists/* \
	&& useradd --system --uid 10001 --home-dir /app --shell /usr/sbin/nologin tas

WORKDIR /app
COPY --from=backend-build /out/tas-server /app/tas-server
COPY --from=frontend-build /src/frontend/dist /app/frontend/dist
COPY config.example.toml /app/config.example.toml

USER tas
EXPOSE 2005

ENTRYPOINT ["/app/tas-server"]
