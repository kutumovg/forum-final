FROM golang:latest

WORKDIR /app

COPY . .

RUN go mod tidy
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o main .

EXPOSE 80
EXPOSE 443

CMD ["./main"]
