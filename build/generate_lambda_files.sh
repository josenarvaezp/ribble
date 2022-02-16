#!/bin/sh

go build -ldflags "-s -w" -o ./build/lambda_gen/"$1"/gen_job  "$2"