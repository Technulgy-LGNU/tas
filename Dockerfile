# Build the Go application
FROM golang:1.23.3-alpine AS builder-go

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY src/ ./src/

RUN go build ./src/main.go

# Final image
FROM alpine:latest

RUN apk --no-cache add ca-certificates wget

WORKDIR /app

COPY --from=builder-go /app/main .

EXPOSE 80

CMD ["./main", "prod"]
