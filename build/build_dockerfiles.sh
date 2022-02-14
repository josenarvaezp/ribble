#!/bin/sh

docker build -t "$1" -f ./build/lambda_gen/"$2"/dockerfiles/Dockerfile."$3" . 