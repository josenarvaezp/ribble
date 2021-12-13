# displ

## Build image
docker build -t mapper .

## Test locally 
mkdir -p ~/.aws-lambda-rie && curl -Lo ~/.aws-lambda-rie/aws-lambda-rie \
https://github.com/aws/aws-lambda-runtime-interface-emulator/releases/latest/download/aws-lambda-rie-arm64  \
&& chmod +x ~/.aws-lambda-rie/aws-lambda-rie    

docker run -d -v ~/.aws-lambda-rie:/aws-lambda --entrypoint /aws-lambda/aws-lambda-rie -p 9000:8080 mapper /lambdas/mapper     

curl -XPOST "http://localhost:9000/2015-03-31/functions/function/invocations" -d '{"bucket": "jobbucket", "id": "fa44fa0d-89de-416f-badc-00e2300acad2", "queues": 10, "size": 7375844,"rangeObjects": [{"objectBucket":"mybucket", "objectKey":"file1.csv","initialByte": 0,"finalByte": 1494763}, {"objectBucket":"mysecondbucket","objectKey":"file3.csv","initialByte": 0,"finalByte": 5881081}]}'

To bring up localstack:
docker-compose up -d
docker-compose down 
docker-compose build

Create a bucket:
awslocal s3 mb s3://jobbucket
awslocal s3 mb s3://mybucket
awslocal s3 mb s3://mysecondbucket
awslocal s3 mb s3://09cd3797-1b53-4c61-b24f-b454bbec73a7

Copy csv objects to bucket
awslocal s3 cp ./build/test_data/business-operations-survey-2020-covid-19-csv.csv  s3://mybucket/file1.csv
awslocal s3 cp ./build/test_data/annual-enterprise-survey-2020-financial-year-provisional-size-bands-csv.csv  s3://mybucket/file2.csv

awslocal s3 cp ./build/test_data/annual-enterprise-survey-2020-financial-year-provisional-csv.csv s3://mysecondbucket/file3.csv 

awslocal s3api get-object --bucket jobbucket --key metadata/fa44fa0d-89de-416f-badc-00e2300acad2 test

// add config to job bucket
awslocal s3 cp ./config.yaml s3://09cd3797-1b53-4c61-b24f-b454bbec73a7/config.yaml