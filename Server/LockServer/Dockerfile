# syntax=docker/dockerfile:1

FROM golang:alpine
#insall gcc
RUN apk add build-base

ADD . /go/src/myapp
WORKDIR /go/src/myapp/LockServer


RUN go mod download
#arg not env so can be overriden by docker compose
#ARG PORT=8888   
#ENV PORT=8888

#RUN go env -w GO111MODULE=off
#RUN go build
EXPOSE 8888

ENTRYPOINT [ "go", "run", "-race", "lockServer.go"]