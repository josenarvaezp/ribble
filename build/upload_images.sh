#!/bin/sh

docker tag "$1":"$2" "$3".dkr.ecr."$4".amazonaws.com/"$1":"$2"
docker push "$2".dkr.ecr."$3".amazonaws.com/"$4"