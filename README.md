# displ

## Build image
docker build -t mapper .

## Test locally 
mkdir -p ~/.aws-lambda-rie && curl -Lo ~/.aws-lambda-rie/aws-lambda-rie \
https://github.com/aws/aws-lambda-runtime-interface-emulator/releases/latest/download/aws-lambda-rie-arm64  \
&& chmod +x ~/.aws-lambda-rie/aws-lambda-rie    

docker run -d -v ~/.aws-lambda-rie:/aws-lambda --entrypoint /aws-lambda/aws-lambda-rie -p 9000:8080 mapper /lambdas/mapper     

curl -XPOST "http://localhost:9000/2015-03-31/functions/function/invocations" -d '{"bucket": "jobbucket", "id": "fa44fa0d-89de-416f-badc-00e2300acad2", "queues": 10, "size": 2713576,"rangeObjects": [{"objectBucket":"mybucket", "objectKey":"file1.csv","initialByte": 0,"finalByte": 1356788}, {"objectBucket":"mysecondbucket","objectKey":"file3.csv","initialByte": 0,"finalByte": 1356788}]}'


curl -XPOST "http://localhost:9000/2015-03-31/functions/function/invocations" -d '{"bucket": "jobbucket", "id": "fa44fa0d-89de-416f-badc-00e2300acad2", "queues": 5, "size": 2713576,"rangeObjects": [{"objectBucket":"mybucket", "objectKey":"file1.csv","initialByte": 0,"finalByte": 1356788}, {"objectBucket":"mysecondbucket","objectKey":"file3.csv","initialByte": 0,"finalByte": 1356788}]}'

// current working mapper invokation
curl -XPOST "http://localhost:9000/2015-03-31/functions/function/invocations" -d '{"jobID": "9499f791-e983-4242-96bb-adeaddb84a51", "mapping": {"bucket": "jobbucket", "id": "fa44fa0d-89de-416f-badc-00e2300acad2", "queues": 5, "size": 2713576,"rangeObjects": [{"objectBucket":"mybucket", "objectKey":"file1.csv","initialByte": 0,"finalByte": 1356788}, {"objectBucket":"mysecondbucket","objectKey":"file3.csv","initialByte": 0,"finalByte": 1356788}]}}'

// current mapper working invokation
curl -XPOST "http://localhost:9001/2015-03-31/functions/function/invocations" -d '{"jobID": "39a0d128-3f80-49a0-83a3-e49ed424c099", "jobBucket": "jobbucket", "QueueName": "-dlq"}'

// reducer
docker run -d -v ~/.aws-lambda-rie:/aws-lambda --entrypoint /aws-lambda/aws-lambda-rie -p 9001:8081 reducer /lambdas/reducer 



To bring up localstack:
docker-compose up -d
docker-compose down 
docker-compose build

Create a bucket:
awslocal s3 mb s3://jobbucket
awslocal s3 mb s3://mybucket
awslocal s3 mb s3://mysecondbucket

Copy csv objects to bucket
awslocal s3 cp ./build/test_data/annual-enterprise-survey-2020-financial-year-provisional-size-bands-csv.csv  s3://mybucket/file1.csv
awslocal s3 cp ./build/test_data/annual-enterprise-survey-2020-financial-year-provisional-size-bands-csv.csv  s3://mysecondbucket/file3.csv


awslocal s3 ls s3://jobbucket/metadata/
awslocal s3api get-object --bucket jobbucket --key metadata/fa44fa0d-89de-416f-badc-00e2300acad2 test

awslocal s3api get-object --bucket jobbucket --key output/feff263f-9da5-4ad9-af60-e2edf5292828 ro5

// add config to job bucket
awslocal s3 cp ./config.yaml s3://09cd3797-1b53-4c61-b24f-b454bbec73a7/config.yaml


13330aa9-d31a-4638-b870-9fb8f8e562cd
417c2069-191f-40a9-b71d-4327a4d3deb3
63d6683a-67f5-449e-9157-b7dda0f77a2f
b8fd0b46-4ed5-41d7-8f4c-3b3fdb22eb63
 feff263f-9da5-4ad9-af60-e2edf5292828
