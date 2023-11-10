FROM golang:1.21.4-alpine3.18

WORKDIR /src/client

COPY go.mod .

RUN go mod download
COPY . .
RUN go build -o /client

CMD ["/client"]