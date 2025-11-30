FROM golang:1.24-alpine

WORKDIR /app

COPY . .

COPY product-service/migrations/ /app/migrations/

COPY product-service/entrypoint.sh .

RUN go build -o product-app ./product-service/main.go

RUN go build -o migrate-app ./product-service/migrate/main.go

RUN chmod +x product-app migrate-app

EXPOSE 8082

CMD ["./entrypoint.sh"]