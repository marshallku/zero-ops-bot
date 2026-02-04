FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o bot ./cmd/bot

FROM alpine:3.20

RUN apk --no-cache add ca-certificates

WORKDIR /app
COPY --from=builder /app/bot .
RUN mkdir -p /app/notes/daily /app/notes/categories && \
    chown -R nobody:nobody /app

USER nobody:nobody

ENTRYPOINT ["./bot"]
