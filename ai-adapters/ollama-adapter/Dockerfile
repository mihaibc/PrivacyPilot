# ---- Build Stage ----
    FROM golang:1.22-alpine AS builder

    WORKDIR /build
    
    COPY go.mod go.sum ./
    RUN go mod download
    
    COPY . .
    
    # Build the application
    # Adjust module path if necessary
    RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/ollama-adapter ./main.go
    
    # ---- Runtime Stage ----
    FROM alpine:latest
    
    WORKDIR /app
    
    COPY --from=builder /app/ollama-adapter /app/ollama-adapter
    
    # Expose the port (default 8084)
    EXPOSE 8084
    
    # Run the application
    ENTRYPOINT ["/app/ollama-adapter"]