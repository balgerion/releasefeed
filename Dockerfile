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
EXPOSE 8080
CMD ["./releasefeed"]