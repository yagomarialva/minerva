# Build stage
FROM golang:1.26-bookworm AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o minerva-backend ./cmd/server

# Final stage
FROM debian:bookworm-slim

# Install Chromium and dependencies for headless browser operations & SQLite
RUN apt-get update && apt-get install -y \
    chromium \
    fonts-liberation \
    ca-certificates \
    sqlite3 \
    libsqlite3-dev \
    calibre \
    && rm -rf /var/lib/apt/lists/*

# Set environment variable for chromedp
ENV CHROME_BIN=/usr/bin/chromium

WORKDIR /app
COPY --from=builder /app/minerva-backend /app/minerva-backend

# Ensure directories exist and set permissions
RUN mkdir -p /app/data /app/books

EXPOSE 8080
CMD ["/app/minerva-backend"]
