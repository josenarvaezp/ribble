#!/bin/sh

docker build --pull --force-rm -t map_sum_aggregator --target map_sum -f ./build/aggregators/Dockerfile .
docker build --pull --force-rm -t map_max_aggregator --target map_max -f ./build/aggregators/Dockerfile .
docker build --pull --force-rm -t map_min_aggregator --target map_min -f ./build/aggregators/Dockerfile .
docker build --pull --force-rm -t sum_aggregator --target sum -f ./build/aggregators/Dockerfile .
docker build --pull --force-rm -t sum_final_aggregator --target sum_final -f ./build/aggregators/Dockerfile .