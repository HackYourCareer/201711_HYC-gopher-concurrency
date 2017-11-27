#!/usr/bin/env bash

#TODO
docker container prune -f
docker run --name ddos-site-instance -d --net=ddos-network --net-alias=mywebsite ddos_site -middleware global -connections 5000