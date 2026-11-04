FROM golang:1.25.11-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/tcgstats ./cmd

FROM alpine:3.22

RUN apk add --no-cache ca-certificates \
    && addgroup -S app \
    && adduser -S -G app app

COPY --from=builder /out/tcgstats /usr/local/bin/tcgstats

USER app
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/tcgstats"]
