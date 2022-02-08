# syntax=docker/dockerfile:1

FROM golang:alpine
#insall gcc
RUN apk add build-base

ADD . /go/src/myapp
WORKDIR /go/src/myapp/Worker


RUN go mod download


#RUN go env -w GO111MODULE=off
#RUN go build
EXPOSE 8888

ENTRYPOINT [ "go", "run", "-race", "worker.go", "8888", "docker"]