FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY . .

RUN cd auth-service && go build -o auth-service .

EXPOSE 8081

CMD ["./auth-service"]