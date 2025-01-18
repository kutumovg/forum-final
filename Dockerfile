FROM golang:1.20-alpine

WORKDIR /app

COPY . .

RUN go mod tidy
RUN go build -o main .

EXPOSE 80
EXPOSE 443

CMD ["./main"]
