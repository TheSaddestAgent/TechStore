FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY . .

RUN go build -o api-gateway-app ./api-gateway

EXPOSE 8080

RUN chmod +x api-gateway-app

CMD ["./api-gateway-app"]