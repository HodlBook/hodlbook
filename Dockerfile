FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o hodlbook ./cmd/main.go

FROM alpine:3.21

WORKDIR /app

RUN apk add --no-cache ca-certificates && \
    adduser -D -u 1000 hodlbook && \
    mkdir -p /data && \
    chown hodlbook:hodlbook /data

COPY --from=builder /app/hodlbook .
COPY --from=builder /app/internal/ui/static ./internal/ui/static
COPY --from=builder /app/internal/ui/templates ./internal/ui/templates
COPY --from=builder /app/docs ./docs

ENV APP_PORT=2008
ENV DB_PATH=/data/hodlbook.db

EXPOSE 2008

VOLUME ["/data"]

USER hodlbook

CMD ["./hodlbook"]
