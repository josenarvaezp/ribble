#!/bin/sh

docker build --pull --force-rm -t map_sum_aggregator --target map_sum -f ./build/aggregators/Dockerfile .