# syntax=docker/dockerfile:1.4
FROM golang:1.24 AS builder

ENV GOPRIVATE=github.com/software-architecture-proj/*

WORKDIR /app

RUN --mount=type=secret,id=github_token \
    git config --global url."https://$(cat /run/secrets/github_token):x-oauth-basic@github.com/".insteadOf "https://github.com/"

COPY go.mod go.sum ./
RUN --mount=type=secret,id=github_token go mod download

COPY . .

RUN go build -o transactions ./main.go

FROM ubuntu:latest

ENV DEBIAN_FRONTEND=noninteractive
ENV TZ=America/Bogota

RUN apt-get update && \
    apt-get install -y tzdata wget curl && \
    ln -fs /usr/share/zoneinfo/$TZ /etc/localtime && \
    dpkg-reconfigure --frontend noninteractive tzdata && \
    wget https://github.com/fullstorydev/grpcurl/releases/download/v1.9.3/grpcurl_1.9.3_linux_amd64.deb && \
    dpkg -i grpcurl_1.9.3_linux_amd64.deb

WORKDIR /root/

COPY --from=builder /app/transactions .

EXPOSE 50051

CMD ["./transactions"]
