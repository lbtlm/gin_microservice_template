# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install git (needed for go mod download)
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o auth-service cmd/server/main.go

# Final stage
FROM alpine:latest

# Install ca-certificates and timezone data
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/auth-service .

# Copy any necessary config files
COPY --from=builder /app/api/swagger ./docs

# Create logs directory
RUN mkdir -p Logs

# Expose port
EXPOSE 8083

# Command to run
CMD ["./gin_middleware_oss"]
