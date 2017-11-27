#!/usr/bin/env bash

# normal usage
./site  -middleware leaky -connections 5000 -timeout 500ms \
    -maxInactiveClientTime 10s  -connPerUser 500 -refillPeriod 100ms


# leaky test
#./site  -middleware leaky -connections 100 -timeout 100ms \
#    -maxInactiveClientTime 10s  -connPerUser 10 -refillPeriod 100ms


# removeINactiveClients:
#  -maxInactiveClientTime 5ms
# -refillPeriod 10ms
#./site  -middleware leaky -connections 5000 -timeout 500ms \
#    -maxInactiveClientTime 5ms  -connPerUser 500 -refillPeriod 10ms &> inactive.txt

