# Use Go 1.24 as base image
FROM golang:1.24-alpine AS builder

# Install FFmpeg and other dependencies
RUN apk add --no-cache ffmpeg git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY src/go.mod src/go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY src/ ./

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o whatsapp .

# Final stage
FROM alpine:latest

# Install FFmpeg and other runtime dependencies
RUN apk add --no-cache ffmpeg ca-certificates tzdata sqlite

# Create app directory
WORKDIR /app

# Create necessary directories
RUN mkdir -p storages statics/media statics/qrcode statics/senditems

# Copy the binary from builder stage
COPY --from=builder /app/whatsapp .

# Copy static files if they exist
COPY --from=builder /app/statics ./statics/ || true
COPY --from=builder /app/views ./views/ || true

# Set permissions
RUN chmod +x whatsapp

# Expose port
EXPOSE 3000

# Set environment variables
ENV APP_PORT=3000
ENV APP_DEBUG=false

# Run the application in REST mode
CMD ["./whatsapp", "rest"]