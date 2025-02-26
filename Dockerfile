# Build stage
FROM golang:1.21 AS builder

WORKDIR /app
COPY . .

# Download dependencies
RUN go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /app/bin/base .

# Final stage
FROM alpine:latest

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/bin/base /app/base

# Create a non-root user
RUN adduser -D base
USER base

# Expose the application port
EXPOSE 16000

# Run the application
CMD ["./base"]