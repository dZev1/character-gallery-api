FROM golang:1.26.4-alpine AS builder
WORKDIR /app

COPY src/go.mod src/go.sum ./
RUN go mod download

COPY src/ .

RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/main .

COPY openapi.yaml /openapi.yaml

ENV DATABASE_URL=""
ENV REDIS_URL="redis://localhost:6379/0"

EXPOSE 8080
CMD ["./main"]
