# Upload file sample

This is a example to upload and download file by golang

## Build a http server by docker

Run command
```
docker build . -t 'http-golang'
docker run -d --name upload-server --restart always -p 80:80 -p 443:443 -v $PWD/Downloads:/server/files -d http-golang
```

## certificate
Its certificate build by [mkcert](https://github.com/FiloSottile/mkcert) tools 

# to-do list
- Use restful API
