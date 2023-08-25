FROM golang:1.19.8-alpine3.16

WORKDIR /app

COPY . .

RUN go build -o main ./cmd

CMD ["./main"]
