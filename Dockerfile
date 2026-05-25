FROM golang:1.26.3-bookworm AS builder
RUN apt-get update && apt-get install -y make git && rm -rf /var/lib/apt/lists/*

RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.30.0
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

WORKDIR /app

COPY . .

RUN make prod

FROM debian:bookworm-slim
WORKDIR /app

RUN apt-get update && apt-get install -y ca-certificates postgresql-client curl && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/bin/gama /app/gama
COPY --from=builder /go/bin/goose /usr/local/bin/goose

RUN curl -fsSL https://dl.min.io/client/mc/release/linux-amd64/mc -o /usr/local/bin/mc && \
    chmod +x /usr/local/bin/mc

COPY config.yaml .
COPY .env .

COPY internal/db/migrations ./migrations
COPY scripts/entrypoint.sh ./entrypoint.sh

RUN sed -i 's/\r$//' ./entrypoint.sh && chmod +x ./entrypoint.sh

EXPOSE 8080
ENTRYPOINT ["./entrypoint.sh"]
