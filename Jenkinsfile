pipeline {
    agent {label '!master'}
    stages { 
        stage('Build') {
            steps {
                sh "docker rm -f new-server || true"
                sh 'go build -o server ./src/app/'
                sh "docker pull ubuntu:18.04"
                sh "docker build . -t newserver:dev"
            }
        }
        stage('Run') {
            steps {
                sh "docker run -d --restart always -p 80:80 -p 443:443 -v /home/junbo/Downloads/:/app/files --name new-server -w /app -v $WORKSPACE:/app newserver:dev /bin/bash -c 'mkdir -p /app/files & (redis-server) & (sleep 10; /app/server)'"
            }
        }
    }
}