#!/bin/sh

aws ecr get-login-password --region "$4" | docker login --username AWS --password-stdin "$3".dkr.ecr."$4".amazonaws.com
docker tag "$1":"$2" "$3".dkr.ecr."$4".amazonaws.com/"$1":"$2"
docker push "$3".dkr.ecr."$4".amazonaws.com/"$1":"$2"