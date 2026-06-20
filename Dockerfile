FROM golang:1.26.0-alpine AS builder
WORKDIR /app

COPY src/go.mod src/go.sum ./
RUN go mod download

COPY src/ .

RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

RUN mkdir -p /root/internal/database/postgres_gallery/

COPY --from=builder /app/main .

COPY src/internal/database/postgres_gallery/schema.sql /root/internal/database/postgres_gallery/schema.sql

COPY src/item_pool.json .

ENV REDIS_URL=localhost:6379
ENV REDIS_PASSWORD=""

EXPOSE 8080
CMD ["./main"]