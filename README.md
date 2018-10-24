# Upload file sample

This is a example to upload and download file by golang & nginx

## Build a http server by docker

Run command
```
docker build . -t 'http-golang'
docker run -d --name upload-server --restart always -p 80:80 -p 9999:9999 -v $PWD/Downloads:/var/www/html/Downloads -d http-golang
```

## Upload file
Open a browser and input http://<your server ip>:<port>/Downloads

## Download file
Open a browser and input http://<your server ip>:<port>/upload.html

# to-do list
- Enable HTTPS
- Use restful API
