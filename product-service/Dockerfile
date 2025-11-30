FROM golang:1.24-alpine

WORKDIR /app

COPY . .

RUN cd product-service && go build -o /product-service .

EXPOSE 8082

CMD ["./product-service"]