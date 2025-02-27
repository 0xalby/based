# Build stage
FROM golang:1.21 AS builder
WORKDIR /api
COPY . .

# Download dependencies
RUN go mod download

# Build the application
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /api/bin/based .

# Final stage
FROM alpine:latest
WORKDIR /api

# Copy the binary from the builder stage
COPY --from=builder /api/bin/based /app/based

# Create a non-root user
RUN adduser -D based
USER based

# Expose the application port
EXPOSE 16000

# Run the application
CMD ["./based"]