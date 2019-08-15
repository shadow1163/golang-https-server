#!/bin/bash

redis-server &
sleep 5
go run /server/main.go