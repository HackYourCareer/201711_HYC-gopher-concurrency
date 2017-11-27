#!/usr/bin/env bash

cd ./../cmd/site

env GOOS=linux go build -o site
docker build -t ddos_site .