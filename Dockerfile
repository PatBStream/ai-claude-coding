# Dockerfile
# Build stage
FROM golang:1.21-alpine AS builder

# Add necessary build tools
RUN apk add --no-cache git gcc musl-dev

# Set working directory
WORKDIR /app

# Initialize Go module
COPY main.go .
RUN go mod init high-performance-server

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server .

# Final stage
FROM alpine:3.19

# Add non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Add necessary runtime dependencies
RUN apk add --no-cache ca-certificates tzdata netcat-openbsd

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/server .

# Set ownership
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 28999

# Command to run
CMD ["./server", "-port", "28999"]


# .dockerignore
#.git
#.gitignore
#README.md
#*.md
#.env
#*.log
#tmp/
#.DS_Store