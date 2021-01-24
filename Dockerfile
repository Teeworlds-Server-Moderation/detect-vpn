FROM golang:alpine as build

LABEL maintainer "github.com/jxsl13"

WORKDIR /build
COPY *.go ./
COPY go.* ./

ENV CGO_ENABLED=0
ENV GOOS=linux 

RUN go get -d && go build -a -ldflags '-w -extldflags "-static"' -o vpn-detection .


FROM alpine:latest as minimal

ENV BROKER_ADDRESS=tcp://mosquitto:1883
ENV CLIENT_ID=detect-vpn

# unencrypted traffic
EXPOSE 1883
# encrypted traffic
EXPOSE 8883


WORKDIR /app
COPY --from=build /build/vpn-detection .
VOLUME ["/data"]
ENTRYPOINT ["/app/vpn-detection"]