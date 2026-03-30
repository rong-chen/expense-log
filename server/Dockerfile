# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Enable CGO disabled for scratch/alpine compatibility
ENV CGO_ENABLED=0
ENV GOOS=linux

# Install dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .
RUN go build -ldflags="-w -s" -o expense-log-server ./cmd/main.go

# Run stage
FROM alpine:latest

# Add timezone data and CA certificates for IMAP/SSL
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/expense-log-server .

# Copy configs (this handles default config.yaml)
COPY --from=builder /app/configs ./configs

# Ensure attachment dir exists
RUN mkdir -p /app/attachments

EXPOSE 8808

CMD ["./expense-log-server"]
