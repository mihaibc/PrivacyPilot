# ---- Build Stage ----
    FROM golang:1.22-alpine AS builder

    # Set working directory relative to the build context (services/ai-coordinator)
    WORKDIR /build
    
    # Copy module files first for caching
    COPY go.mod go.sum ./
    RUN go mod download
    
    # Copy the rest of the source code
    COPY . .
    
    # Build the application
    # Assumes the module path is github.com/your-username/PrivacyPilot/services/ai-coordinator
    # Adjust if necessary
    RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/ai-coordinator ./main.go
    
    # ---- Runtime Stage ----
    FROM alpine:latest
    
    WORKDIR /app
    
    # Copy the binary from the builder stage
    COPY --from=builder /app/ai-coordinator /app/ai-coordinator
    
    # Expose the port (default 8083)
    EXPOSE 8083
    
    # Run the application
    ENTRYPOINT ["/app/ai-coordinator"]