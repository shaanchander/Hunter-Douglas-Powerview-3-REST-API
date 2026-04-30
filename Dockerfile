FROM golang:1.26-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/powerview-api ./cmd/api

FROM alpine:3.22

RUN addgroup -S app && adduser -S -G app app
WORKDIR /app

COPY --from=builder /out/powerview-api /usr/local/bin/powerview-api

USER app
EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/powerview-api"]