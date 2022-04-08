#!/bin/sh

awslocal s3 ls

[ $? == 0 ] || fail 1 "Localstack not live..."