FROM golang:1.20-bullseye

RUN mkdir /app
RUN mkdir /data

WORKDIR /app

COPY src /app

RUN cd /app
RUN go build -o ./cmd/node/node ./cmd/node/
RUN go build -o ./cmd/dns/dns ./cmd/dns/
RUN go build -o ./cmd/wallet/wallet ./cmd/wallet/

ENV GO111MODULE=on
ENV GOFLAGS=-mod=vendor

