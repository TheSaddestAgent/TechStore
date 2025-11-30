FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY . .

RUN go build -o auth-app ./auth-service

EXPOSE 8081

RUN chmod +x auth-app

CMD ["./auth-app"]