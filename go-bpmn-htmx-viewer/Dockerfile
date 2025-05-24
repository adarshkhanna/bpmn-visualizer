# Stage 1: Build the Go application
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum first to leverage Docker cache
COPY go.mod go.sum ./
RUN go mod download
RUN go mod verify

# Copy the rest of the application source code
COPY . .

# Build the application
# -ldflags="-w -s" reduces the size of the binary by removing debug information
# CGO_ENABLED=0 ensures a static binary without C dependencies
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /server cmd/server/main.go

# Stage 2: Create the final lightweight image
FROM alpine:latest

WORKDIR /app

# Copy static assets and templates
# Ensure these paths match your project structure and where the Go app expects them
COPY --from=builder /app/static ./static
COPY --from=builder /app/templates ./templates

# Copy the compiled binary from the builder stage
COPY --from=builder /server .

# Expose the port the application will run on.
# This should match the port your application listens on (default 8080, or from PORT env var)
EXPOSE 8080

# Set the PORT environment variable that the app uses
ENV PORT=8080

# Command to run the application
# The binary is now at /app/server in this stage
CMD ["/app/server"]
