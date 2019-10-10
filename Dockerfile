FROM ubuntu:18.04

MAINTAINER shadow1163 (674602286@qq.com)

RUN sed -i "s@archive.ubuntu.com@mirrors.aliyun.com@g" /etc/apt/sources.list &&\
    sed -i "s@/security.ubuntu.com/@/mirrors.aliyun.com/@g" /etc/apt/sources.list &&\
    apt-get update && apt-get install -y redis-server