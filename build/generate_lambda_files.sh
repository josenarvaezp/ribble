#!/bin/sh

go build -o ./build/lambda_gen/"$1"/gen_job  "$2"