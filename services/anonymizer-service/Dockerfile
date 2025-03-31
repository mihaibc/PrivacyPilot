# ---- Build Stage ----
    FROM golang:1.22-alpine AS builder

    WORKDIR /build
    
    COPY go.mod go.sum ./
    RUN go mod download
    
    COPY . .
    
    # Build the application
    RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/anonymizer-service ./main.go
    
    # ---- Runtime Stage ----
    FROM alpine:latest
    
    WORKDIR /app
    
    # Copy the binary from the builder stage
    COPY --from=builder /app/anonymizer-service /app/anonymizer-service
    
    # Expose the port (default 8081)
    EXPOSE 8081
    
    # Run the application
    ENTRYPOINT ["/app/anonymizer-service"]