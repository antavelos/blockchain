FROM golang:1.20-bullseye

RUN go install github.com/antavelos/blockchain/cmd/node@latest

ENV GO111MODULE=on
ENV GOFLAGS=-mod=vendor

