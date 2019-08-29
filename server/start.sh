#!/bin/bash

redis-server &
sleep 5
cd /server
go run main.go