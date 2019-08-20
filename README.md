# build a https server using golang language

This is a example to upload and download file by golang

## Build a http server by docker

Run command
```
docker build . -t 'http-golang'
docker run -d --name upload-server --restart always -p 80:80 -p 443:443 -v $PWD/Downloads:/server/files -d http-golang
```

## certificate
Its certificate build by [mkcert](https://github.com/FiloSottile/mkcert) tools 

## key point
- path /: normal http file server and main page. Using html + css + js.
- path /key: a sample key mini game. Using html + canvas.
- path /chatroom: a sample chat room. Using html + vue.js + websocket.
- path /api/notes: a smaple restful API. Using golang.