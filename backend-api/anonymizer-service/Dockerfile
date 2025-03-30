FROM golang:1.18-alpine
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o anonymizer-service main.go
EXPOSE 8081
CMD ["./anonymizer-service"]
