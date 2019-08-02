FROM ubuntu:18.04

MAINTAINER shadow1163 (674602286@qq.com)

RUN apt-get update && apt-get install -y software-properties-common

RUN add-apt-repository -y ppa:longsleep/golang-backports

RUN apt-get update && apt-get install -y golang

COPY server /server

RUN mkdir /server/files

EXPOSE 80 443

CMD go run /server/main.go
