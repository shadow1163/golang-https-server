FROM ubuntu:18.04

MAINTAINER shadow1163 (674602286@qq.com)

RUN sed -i "s@archive.ubuntu.com@cn.archive.ubuntu.com@g" /etc/apt/sources.list && apt-get update && apt-get install -y software-properties-common

RUN add-apt-repository -y ppa:longsleep/golang-backports && sed -i "s/ppa\.launchpad\.net/lanuchpad.moruy.cn/g" /etc/apt/sources.list.d/*.list

RUN apt-get update && apt-get install -y --fix-missing golang git redis-server wget unzip

RUN go get github.com/gorilla/mux \
    && go get github.com/gomodule/redigo/redis \
    && go get github.com/satori/go.uuid \
    && go get github.com/gorilla/websocket \
    && go get -u github.com/golang/protobuf/protoc-gen-go

RUN mkdir /root/go/src/google.golang.org \
    && mkdir -p /root/go/src/golang.org/x \
    && (go get -u github.com/grpc/grpc-go; exit 0) \
    && ln -s /root/go/src/github.com/grpc/grpc-go/ /root/go/src/google.golang.org/grpc \
    && (go get github.com/google/go-genproto; exit 0) \
    && ln -s /root/go/src/github.com/google/go-genproto /root/go/src/google.golang.org/genproto \
    && go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway \
    && (go get github.com/golang/net; exit 0) \
    && ln -s /root/go/src/github.com/golang/net /root/go/src/golang.org/x/net \
    && (go get github.com/golang/sys; exit 0) \
    && ln -s /root/go/src/github.com/golang/sys /root/go/src/golang.org/x/sys \
    && (go get github.com/golang/text; exit 0) \
    && ln -s /root/go/src/github.com/golang/text /root/go/src/golang.org/x/text \
    && go get -u github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger


## install protobuf
RUN wget https://github.com/google/protobuf/releases/download/v3.9.1/protobuf-all-3.9.1.zip \
    && unzip protobuf-all-3.9.1.zip \
    && cd protobuf-3.9.1/ \
    && ./configure \
    && make \
    && make install \
    && ldconfig \
    && cd ..

COPY server /server

COPY swagger-ui /swagger-ui

COPY note /note

RUN mkdir /server/files

EXPOSE 80 443 50051

CMD bash /server/start.sh
