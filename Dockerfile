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
ENV REDIS_ADDRESS=redis:6379
ENV REDIS_PASSWORD=""
ENV REDIS_DB=1
ENV DATA_PATH=/data
ENV BLACKLIST_FOLDER=blacklists
ENV WHITELIST_FOLDER=whitelists
ENV VPN_BAN_REASON="VPN - https://zcat.ch/bans"
# can be 10m, 12s, 12h, 1h12m, 1h12m13s
ENV VPN_BAN_DURATION=24h
ENV BROADCAST_BANS=false
ENV DEFAULT_BAN_COMMAND="ban {IP} {DURATION:MINUTES} {REASON}"


WORKDIR /app
COPY --from=build /build/vpn-detection .
VOLUME ["/data"]
ENTRYPOINT ["/app/vpn-detection"]