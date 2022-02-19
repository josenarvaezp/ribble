#!/bin/sh

docker tag "$1":"$2" "$3".dkr.ecr."$4".amazonaws.com/"$1":"$2"
docker push "$3".dkr.ecr."$4".amazonaws.com/"$1":"$2"