#!/bin/sh

echo "Building lambda..."
# mkdir "$1"/localstack
# chmod 0755 "$1"/localstack

# GOOS=linux go build -o "$1"/localstack "$2" && \
#     zip "$1"/localstack/"$3".zip "$1"/localstack/"$3"

echo "Creating lambda function..."
awslocal lambda create-function \
    --region "$4" \
    --function-name "$5" \
    --runtime go1.x \
    --handler "$3" \
    --timeout 60 \
    --memory-size 128 \
    --zip-file fileb://"$1"/localstack/"$3".zip \
    --role arn:aws:iam::123456:role/test-role


#  within the dir
1. go mod init WordCount
2. go mod tidy
3. GOOS=linux go build 
4. zip WordCount.zip WordCount
5. awslocal lambda create-function \
    --region eu-west-2 \
    --function-name wordcount_05c3d08f-b08a-4247-af66-682b49631443_3 \
    --runtime go1.x \
    --handler WordCount \
    --timeout 60 \
    --memory-size 128 \
    --zip-file fileb://build/lambda_gen/05c3d08f-b08a-4247-af66-682b49631443/map/WordCount.zip \
    --role arn:aws:iam::123456:role/test-role
6. awslocal lambda invoke \
    --region eu-west-2 \
    --function-name wordcount_05c3d08f-b08a-4247-af66-682b49631443_3  \
    --invocation-type Event \
    --cli-binary-format raw-in-base64-out \
    --payload '{ "Name": "Jose"}' \
    out.out


#  GOOS=linux go build -o ./build/lambda_gen/8c8177a1-c962-4b4b-876a-4a85f9e8b9f6/map/ ./build/lambda_gen/8c8177a1-c962-4b4b-876a-4a85f9e8b9f6/map/WordCount.go   && \
#     zip ./build/lambda_gen/8c8177a1-c962-4b4b-876a-4a85f9e8b9f6/map/WordCount.zip ./build/lambda_gen/8c8177a1-c962-4b4b-876a-4a85f9e8b9f6/map/WordCount

# awslocal lambda create-function \
#     --region eu-west-2 \
#     --function-name wordcount_8c8177a1-c962-4b4b-876a-4a85f9e8b9f6 \
#     --runtime go1.x \
#     --handler WordCount \
#     --timeout 60 \
#     --memory-size 128 \
#     --zip-file fileb://build/lambda_gen/8c8177a1-c962-4b4b-876a-4a85f9e8b9f6/map/WordCount.zip \
#     --role arn:aws:iam::123456:role/test-role

# awslocal lambda invoke \
#     --region eu-west-2 \
#     --function-name wordcount_05c3d08f-b08a-4247-af66-682b49631443  \
#     --invocation-type Event \
#     --cli-binary-format raw-in-base64-out \
#     --payload '{ "Name": "Jose"}' \
#     out.out


# #!/bin/sh

# echo "Building lambda..."
# mkdir "$1"/localstack
# chmod 0755 "$1"/localstack

# # IFS='.'; COMPLETE_GENERATED_FILE=($2); unset IFS;
# # REPO_URI=(${REPO_URI_NAME[2]})
# # REPO_URI=${REPO_URI%??}
# # REPO_URI="localhost:$REPO_URI"




# GOOS=linux go build -o "$1"/localstack "$2" && \
#     zip "$1"/localstack/"$3".zip "$1"/localstack/"$3"

# echo "Creating lambda function..."
# awslocal lambda create-function \
#     --region "$4" \
#     --function-name "$5" \
#     --runtime go1.x \
#     --handler "$3" \
#     --timeout 60 \
#     --memory-size 128 \
#     --zip-file fileb://"$1"/localstack/"$3".zip \
#     --role arn:aws:iam::123456:role/test-role

# #  GOOS=linux go build -o ./build/lambda_gen/8c8177a1-c962-4b4b-876a-4a85f9e8b9f6/map/ ./build/lambda_gen/8c8177a1-c962-4b4b-876a-4a85f9e8b9f6/map/WordCount.go   && \
# #     zip ./build/lambda_gen/8c8177a1-c962-4b4b-876a-4a85f9e8b9f6/map/WordCount.zip ./build/lambda_gen/8c8177a1-c962-4b4b-876a-4a85f9e8b9f6/map/WordCount

# # awslocal lambda create-function \
# #     --region eu-west-2 \
# #     --function-name wordcount_8c8177a1-c962-4b4b-876a-4a85f9e8b9f6 \
# #     --runtime go1.x \
# #     --handler WordCount \
# #     --timeout 60 \
# #     --memory-size 128 \
# #     --zip-file fileb://build/lambda_gen/8c8177a1-c962-4b4b-876a-4a85f9e8b9f6/map/WordCount.zip \
# #     --role arn:aws:iam::123456:role/test-role

# # awslocal lambda invoke \
# #     --region eu-west-2 \
# #     --function-name wordcount_05c3d08f-b08a-4247-af66-682b49631443  \
# #     --invocation-type Event \
# #     --cli-binary-format raw-in-base64-out \
# #     --payload '{ "Name": "Jose"}' \
# #     out.out