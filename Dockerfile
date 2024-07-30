FROM golang:1.20.12-alpine

WORKDIR /app

COPY . .

RUN go get -u github.com/swaggo/swag/cmd/swag
RUN swag init
RUN go build -o main .

EXPOSE 8080

CMD ["./main"]
