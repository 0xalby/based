# Build stage
FROM golang:latest AS builder
RUN apk update && apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o bin/base -ldflags="-s -w"
# Final stage
FROM alpine:latest
RUN addgroup -S base && adduser -S base -G base
WORKDIR /app
COPY --from=builder /app/bin/base /app/bin/base
RUN chown -R base:base /app
EXPOSE 16000
USER base
CMD ["/app/bin/base"]