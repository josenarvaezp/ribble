#!/bin/sh

docker build -t MapSumAggregator --target map_sum -f ./build/aggregators/Dockerfile .