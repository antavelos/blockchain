FROM golang:1.20-bullseye

RUN mkdir /app
RUN mkdir /data

WORKDIR /app

COPY ./src /app

RUN cd /app
RUN go build -o ./internal/cmd/node/node ./internal/cmd/node/
RUN go build -o ./internal/cmd/dns/dns ./internal/cmd/dns/
RUN go build -o ./internal/cmd/wallet/wallet ./internal/cmd/wallet/
RUN go build -o ./internal/cmd/admin/admin ./internal/cmd/admin/

ENV GO111MODULE=on
ENV GOFLAGS=-mod=vendor

