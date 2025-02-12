# Use the official Go image as the base image
FROM golang:1.23.4-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Use a minimal alpine image for the final stage
FROM alpine:latest

# Set the working directory
WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/main .

# Copy the static files and templates
COPY --from=builder /app/static ./static
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/locales ./locales
COPY --from=builder /app/md ./md
COPY --from=builder /app/assets ./assets

# Expose the port the app runs on
EXPOSE 80

# Add after EXPOSE 80
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["wget", "-q", "--spider", "http://localhost:80/v1/health"]

# Run the binary
CMD ["./main"]