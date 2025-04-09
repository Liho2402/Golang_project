# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./
# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
# Disabling CGO allows for a static binary build
# -ldflags="-w -s" strips debug information and symbols to reduce binary size
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /authservice ./cmd/authservice

# Final stage
FROM alpine:latest

WORKDIR /root/

# Copy the pre-built binary file from the previous stage
COPY --from=builder /authservice .

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./authservice"] 