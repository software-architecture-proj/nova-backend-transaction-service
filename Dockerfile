# syntax=docker/dockerfile:1.4 
FROM golang:1.24 AS builder

ENV GOPRIVATE=github.com/software-architecture-proj/*

WORKDIR /app

RUN --mount=type=secret,id=github_token \
    git config --global url."https://$(cat /run/secrets/github_token):x-oauth-basic@github.com/".insteadOf "https://github.com/"

COPY go.mod ./
RUN --mount=type=secret,id=github_token go mod download


COPY . .
RUN CGO_ENABLED=1 go build -o transactions ./server.go

FROM ubuntu:latest

ENV DEBIAN_FRONTEND=noninteractive
ENV TZ=America/Bogota

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

RUN apt-get update && \
    apt-get install -y tzdata && \
    ln -fs /usr/share/zoneinfo/$TZ /etc/localtime && \
    dpkg-reconfigure --frontend noninteractive tzdata && \
    apt install wget -y && \
    wget https://github.com/fullstorydev/grpcurl/releases/download/v1.9.3/grpcurl_1.9.3_linux_amd64.deb && \
    dpkg -i grpcurl_1.9.3_linux_amd64.deb 

WORKDIR /root/
COPY --from=builder /app/transactions .
COPY --from=builder /app/proto/transaction_service.proto .
CMD ["/entrypoint.sh"]