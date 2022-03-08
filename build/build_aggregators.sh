#!/bin/sh

docker build --pull --force-rm -t map_aggregator --target map_aggregator -f ./build/aggregators/Dockerfile .
docker build --pull --force-rm -t map_random_aggregator --target map_random_aggregator -f ./build/aggregators/Dockerfile .
docker build --pull --force-rm -t final_aggregator --target final_aggregator -f ./build/aggregators/Dockerfile .