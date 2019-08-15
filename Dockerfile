FROM ubuntu:18.04

MAINTAINER shadow1163 (674602286@qq.com)

RUN apt-get update && apt-get install -y software-properties-common

RUN add-apt-repository -y ppa:longsleep/golang-backports

RUN apt-get update && apt-get install -y golang git redis-server

RUN go get github.com/gorilla/mux && go get github.com/gomodule/redigo/redis && go get github.com/satori/go.uuid

COPY server /server

RUN mkdir /server/files

EXPOSE 80 443

CMD bash /server/start.sh
