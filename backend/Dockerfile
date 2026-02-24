# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/server ./cmd/server

# Dev stage (for hot reload)
FROM golang:1.24-alpine AS dev

WORKDIR /app
RUN go install github.com/air-verse/air@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .
CMD ["air", "-c", ".air.toml"]

# Production stage
FROM alpine:3.19 AS production

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /bin/server /app/server
COPY db/migrations /app/migrations

EXPOSE 8080
USER nobody
CMD ["/app/server"]
