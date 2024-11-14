# Use the official Golang image as the builder
FROM golang:1.23.2 AS builder

# Set timezone and install tzdata for timezone information
ENV TZ=Europe/Bucharest
RUN apt-get update && apt-get install -y tzdata

WORKDIR /app

# Copy the go.mod and go.sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire project (including static files for embedding)
COPY . .

# Build the binary with embedded files
RUN CGO_ENABLED=0 go build -o lead-gen-tracker main.go

# Use a minimal alpine image
FROM alpine:latest

# Install tzdata to support timezone configuration
RUN apk add --no-cache tzdata

# Set timezone environment variable in the final stage
ENV TZ=Europe/Bucharest

WORKDIR /

# Copy only the compiled binary from the builder stage
COPY --from=builder /app/lead-gen-tracker /lead-gen-tracker

# Expose port 8080
EXPOSE 8080

# Run the application
CMD ["/lead-gen-tracker"]