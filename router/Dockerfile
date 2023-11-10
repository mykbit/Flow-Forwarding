FROM golang:1.21.4-alpine3.18

WORKDIR /src/router

COPY go.mod .

RUN go mod download
COPY router.go .
RUN go build -o /router

CMD ["/router"]