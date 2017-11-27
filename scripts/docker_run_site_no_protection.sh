#!/usr/bin/env bash


# TODO check how we distinguish users!!!!

#TODO
docker container prune -f
docker run --name ddos-site-instance -d --net=ddos-network --net-alias=mywebsite ddos_site