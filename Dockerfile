FROM ubuntu:18.04

MAINTAINER shadow1163 (674602286@qq.com)

RUN apt-get update && apt-get install -y software-properties-common

RUN add-apt-repository -y ppa:gophers/archive

RUN apt-get update && apt-get install -y golang-1.10-go golang-1.10-doc nginx

RUN ln -s /usr/lib/go-1.10/bin/go /usr/bin/go

COPY main.go /main.go

COPY conf/default /etc/nginx/sites-available/

RUN mkdir /var/www/html/Downloads

COPY upload.html /var/www/html/

RUN rm -f /var/www/html/index.nginx-debian.html

EXPOSE 80 9999

CMD /etc/init.d/nginx restart && go run /main.go
