# ---- Build Stage ----
    FROM golang:1.22-alpine AS builder

    # Set working directory
    WORKDIR /build
    
    # Copy Go module files
    COPY go.mod go.sum ./
    
    # Download dependencies.
    # This leverages Docker cache layers; dependencies are only re-downloaded
    # if go.mod or go.sum change.
    RUN go mod download
    
    # Copy the source code
    COPY . .
    
    # Build the application
    # CGO_ENABLED=0 produces a statically linked binary
    # -ldflags="-w -s" strips debugging information, reducing binary size
    RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/api-gateway ./main.go
    
    # ---- Runtime Stage ----
    FROM alpine:latest
    
    # Set working directory
    WORKDIR /app
    
    # Copy the statically linked binary from the builder stage
    COPY --from=builder /app/api-gateway /app/api-gateway
    
    # Expose the port the application runs on (defined by PORT env var, defaults to 8080)
    # Note: This is documentation; the actual port mapping happens in docker-compose or K8s
    EXPOSE 8080
    
    # Define the command to run the application
    # The binary is executed directly
    ENTRYPOINT ["/app/api-gateway"]