FROM alpine:latest

WORKDIR /traffic

RUN apk update
RUN apk add tcpdump

CMD tcpdump -i any -nn -vvv -w traffic.pcap