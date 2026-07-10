# syntax=docker/dockerfile:1.7

FROM node:26-bookworm AS frontend-build
WORKDIR /src/frontend

COPY frontend/package*.json ./
RUN npm ci

COPY frontend/ ./
RUN npm run build

FROM golang:1.26.4-bookworm AS backend-build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ ./cmd/
COPY internal ./internal
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/tas-server ./cmd/tas-server

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
EXPOSE 8000

ENTRYPOINT ["/app/tas-server"]
