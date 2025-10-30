# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies (gcc, musl-dev needed for sqlite3, nodejs and npm for tailwind)
RUN apk add --no-cache gcc musl-dev nodejs npm

# Set working directory
WORKDIR /app

# Copy package files for npm
COPY package.json package-lock.json tailwind.config.js ./

# Install npm dependencies
RUN npm ci

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build Tailwind CSS
RUN npx tailwindcss -i static/input.css -o static/styles.css --minify

# Build the application
# CGO_ENABLED=1 is required for sqlite3
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o budget-server cmd/server/main.go

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates sqlite

# Create a non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/budget-server .

# Copy static files from builder
COPY --from=builder /app/static ./static

# Create directory for database with proper permissions
RUN mkdir -p /app/data && chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Set environment variables with defaults
ENV PORT=8080
ENV DB_PATH=/app/data/budget.db

# Run the application
CMD ["./budget-server"]
