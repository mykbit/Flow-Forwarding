FROM golang:1.21.4-alpine3.18

WORKDIR /src/endpoint

COPY go.mod .

RUN go mod download
COPY . .
RUN go build -o /endpoint
RUN apk update
RUN apk add tcpdump

CMD tcpdump -i any -w traffic.pcap & /endpoint