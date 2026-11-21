FROM golang:1.26.11-alpine AS builder
WORKDIR /app

COPY src/go.mod src/go.sum ./
RUN go mod download

COPY src/ .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o main ./cmd

FROM alpine:3.21
RUN apk --no-cache add ca-certificates tzdata \
    && adduser -D -H -u 10001 appuser

WORKDIR /app
COPY --from=builder /app/main .
RUN chown appuser:appuser /app/main

USER appuser

ENV DATABASE_URL=""
ENV REDIS_URL="redis://localhost:6379/0"
ENV JWT_SECRET=""
ENV CORS_ALLOW_ORIGIN=""
ENV PORT=":8080"

EXPOSE 8080
CMD ["./main"]
