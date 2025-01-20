# Start with a minimal Go image
FROM golang:1.20-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the app source code
COPY . .

# Build the Go application
RUN go build -o quiz_app

# Create a minimal runtime image
FROM alpine:latest

# Set working directory
WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/quiz_app .

# Expose the port (if your app needs it)
# EXPOSE 8080

# Command to run the application
CMD ["./quiz_app"]
