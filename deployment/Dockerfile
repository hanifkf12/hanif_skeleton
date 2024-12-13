# Stage 1: Build the application
FROM golang:1.23-alpine AS builder

# Set up a working directory
WORKDIR /app

# Install dependencies needed for build
RUN apk add --no-cache gcc g++ libc-dev git

# Copy Go module files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the Go application as a static binary
RUN CGO_ENABLED=0 go build -o app .

# Stage 2: Create a minimal runtime image
FROM alpine:latest

# Set up a working directory
WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /app/app .
COPY database/migration /app/database/migration

# Set environment variables (optional, adjust as needed)
ENV APP_PORT=9000

# Expose the application's port
EXPOSE ${APP_PORT}

# Command to run the application
CMD ["sh", "-c", "./app db:migrate up && ./app http-start"]
