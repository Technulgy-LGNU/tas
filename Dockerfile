# Build the Go application
FROM golang:1.27.1-alpine AS builder-go

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY backend/ ./backend/

RUN go build ./backend/main.go

# Build the Node.js application
FROM node:26.9-alpine AS builder-node

WORKDIR /app

COPY frontend/ ./

RUN npm install

RUN npm run build

# Final image
FROM alpine:latest

RUN apk --no-cache add ca-certificates wget

WORKDIR /app

COPY --from=builder-go /app/main .

COPY --from=builder-node /app/dist ./frontend/dist

EXPOSE 2005

HEALTHCHECK --interval=30s --timeout=5s --start-period=30s --retries=3 \
  CMD curl --fail --silent --show-error --max-time 4 http://127.0.0.1:2005/healthcheck || exit 1

CMD ["./main"]
