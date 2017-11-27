#!/usr/bin/env bash
go build -o attacker

# leaky test
#./attacker -c 20 -repeat 20 -wait 5ms

# show select on default (try to give more tokens that buffer size
#go build -o attacker
#./attacker -c 2 -repeat 100 -wait 5ms


# removeInactive
./attacker -c 200 -repeat 20 -wait 10ms #- removeInactive