FROM golang:1.21.1-alpine3.18

WORKDIR /src/client

COPY go.mod .

RUN go mod download
COPY client.go .
RUN go build -o /client

CMD ["/client"]