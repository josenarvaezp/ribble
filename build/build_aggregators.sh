#!/bin/sh

docker build --pull --force-rm -t map_aggregator --target map_aggregator -f ./build/aggregators/Dockerfile .
# docker build --pull --force-rm -t sum_aggregator --target sum -f ./build/aggregators/Dockerfile .
# docker build --pull --force-rm -t sum_final_aggregator --target sum_final -f ./build/aggregators/Dockerfile .