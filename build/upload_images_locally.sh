#!/bin/sh

BUILD_FILE="$1"
FILE_DIR=${BUILD_FILE%/*}

cd ${FILE_DIR}
go mod init "$2"
go mod tidy
GOOS=linux go build 
zip "$2".zip "$2"

echo "Creating lambda function..."
awslocal lambda create-function \
    --region "$3" \
    --function-name "$4" \
    --runtime go1.x \
    --handler "$2" \
    --timeout 60 \
    --memory-size 128 \
    --zip-file fileb://"$2".zip \
    --role arn:aws:iam::123456:role/test-role