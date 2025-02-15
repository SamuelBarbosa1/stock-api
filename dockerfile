FROM golang:1.23

WORKDIR /app

COPY . .

RUN go mod download

RUN go build -o stock-api .

EXPOSE 8080

CMD ["./stock-api"]