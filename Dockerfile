# Build stage
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

WORKDIR /app

# Copy go mod files and source code
COPY go.mod ./
COPY . .

# Download dependencies and tidy
RUN go mod tidy && go mod download && go mod verify

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bookwise-api ./cmd/server

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/bookwise-api .

# Expose port
EXPOSE 8080

# Run the application
CMD ["./bookwise-api"]

