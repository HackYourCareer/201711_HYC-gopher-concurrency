#!/usr/bin/env bash

docker container prune -f
#docker run --name ddos-site-instance -d --net=ddos-network --net-alias=mywebsite ddos_site \
#    -middleware leaky -connections 5000 -timeout 500ms \
#    -maxInactiveClientTime 60s -connPerUser 500 -refillPeriod 100ms

docker run --name ddos-site-instance -d --net=ddos-network --net-alias=mywebsite ddos_site \
-middleware leaky -connections 100 -timeout 100ms \
    -maxInactiveClientTime 10s  -connPerUser 10 -refillPeriod 100ms