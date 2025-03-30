FROM golang:1.18-alpine
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o api-gateway main.go
EXPOSE 8080
CMD ["./api-gateway"]
