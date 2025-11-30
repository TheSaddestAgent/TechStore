FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY . .

RUN cd api-gateway && go build -o api-gateway .

EXPOSE 8080

CMD ["./api-gateway"]