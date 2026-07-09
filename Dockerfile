FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o releasefeed .

FROM alpine:3.23
WORKDIR /app
RUN apk add --no-cache tzdata
COPY --from=builder /app/releasefeed .
EXPOSE 880
HEALTHCHECK --interval=1m --timeout=5s CMD wget -qO- http://127.0.0.1:880/ || exit 1
CMD ["./releasefeed"]