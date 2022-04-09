#!/bin/sh

awslocal s3 ls

if ! awslocal s3 ls ; then
    echo "Localstack not running..."
    docker-compose up -d

    # wait until is up
    while ! awslocal s3 ls
    do
        echo "Waiting for localstack..."
        sleep 5
    done

    echo "Localstack running..."
else 
    echo "Localstack already running..."
fi