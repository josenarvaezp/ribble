#!/bin/sh

docker build --pull --force-rm -t "$1" -f ./build/lambda_gen/"$2"/dockerfiles/Dockerfile."$3" . 