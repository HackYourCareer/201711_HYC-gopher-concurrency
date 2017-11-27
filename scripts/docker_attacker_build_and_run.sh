#!/usr/bin/env bash

# docker network create ddos-network

cd ./../cmd/attacker/
env GOOS=linux go build -o attacker
docker build -t ddos_attacker .
#docker run -it  --net=ddos-network  ddos_attacker -wait 1s -c 10000 -repeat 1 -endpoint "http://mywebsite:5000/status"

docker run -it  --net=ddos-network  ddos_attacker -c 20 -repeat 20 -wait 5ms -endpoint "http://mywebsite:5000/status"