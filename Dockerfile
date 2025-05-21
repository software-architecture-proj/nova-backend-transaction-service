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

RUN apt-get update && \
    apt-get install -y tzdata && \
    ln -fs /usr/share/zoneinfo/$TZ /etc/localtime && \
    dpkg-reconfigure --frontend noninteractive tzdata

WORKDIR /root/
COPY --from=builder /app/transactions .
CMD ["./transactions"]


# To build the Docker image, run:

#echo "github_pat_xxxxxxxxxxxxxxxxxxx" > ~/.github-token  
#chmod 600 ~/.github-token   
#DOCKER_BUILDKIT=1 docker build \
#  --secret id=github_token,src=$HOME/.github-token \
#  -t transactions:x .

# To run the Docker container, use:
    #docker run --network=host --privileged -it transactions:x