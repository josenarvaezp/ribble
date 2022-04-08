build_cli:
	go build -o ./ribble ./cmd/driver/main.go

lf:
	awslocal s3 ls

setup:
	awslocal s3 mb s3://jobbucket
	awslocal s3 mb s3://mybucket
	awslocal s3 mb s3://mysecondbucket
	awslocal s3 cp ./build/test_data/annual-enterprise-survey-2020-financial-year-provisional-size-bands-csv.csv  s3://mybucket/file1.csv
	awslocal s3 cp ./build/test_data/annual-enterprise-survey-2020-financial-year-provisional-size-bands-csv.csv  s3://mysecondbucket/file3.csv
	go run ./cmd/driver/main.go

make rund:
	go run ./cmd/driver/main.go

runm:
	curl -XPOST "http://localhost:9000/2015-03-31/functions/function/invocations" -d '{"jobID": "1220edd7-6904-44c3-bb35-5679fe3dee74", "mapping": {"id": "fa44fa0d-89de-416f-badc-00e2300acad2", "queues": 5, "size": 2713576,"rangeObjects": [{"objectBucket":"mybucket", "objectKey":"file1.csv","initialByte": 0,"finalByte": 1356788}, {"objectBucket":"mysecondbucket","objectKey":"file3.csv","initialByte": 0,"finalByte": 1356788}]}}'

runr:
	curl -XPOST "http://localhost:9001/2015-03-31/functions/function/invocations" -d '{"jobID": "e9352b6f-3255-47d1-a185-c854722f92a1", "jobBucket": "jobbucket", "queueName": "e9a2c4c7-abb3-4c19-87bc-62e221da864a"}'

runc:
	curl -XPOST "http://localhost:9002/2015-03-31/functions/function/invocations" -d '{"jobID": "e9352b6f-3255-47d1-a185-c854722f92a1", "numQueues": 5, "numMappers": 1}'

buildr:
	docker build -t reducer .

geto:
	awslocal s3api get-object --bucket e33e7151-8ae8-4560-93b2-5ba389949f7d --key output/7ecd2711-56d1-4d01-90a2-90b80e92c7e1 res



runclibuild:
	./ribble build --job  ./examples/wordcount/job/job.go  

remove_images:
	docker images | grep word | awk '{ print $3; }' | xargs docker rmi



set:
	awslocal s3 mb s3://input-bucket
	awslocal s3 cp ./test.txt s3://input-bucket/test.txt

# run build
# run upload

# docker run -d -v ~/.aws-lambda-rie:/aws-lambda --entrypoint /aws-lambda/aws-lambda-rie  -p 9000:8080 coordinator_1220edd7-6904-44c3-bb35-5679fe3dee74:latest /lambdas/coordinator
# curl -XPOST "http://localhost:9000/2015-03-31/functions/function/invocations" -d '{"jobID": "1220edd7-6904-44c3-bb35-5679fe3dee74", "numQueues": 5, "numMappers": 1}'

# docker run -d -v ~/.aws-lambda-rie:/aws-lambda --entrypoint /aws-lambda/aws-lambda-rie  -p 9004:8080 wordcount_1220edd7-6904-44c3-bb35-5679fe3dee74:latest /lambdas/WordCount
# curl -XPOST "http://localhost:9004/2015-03-31/functions/function/invocations" -d '{"jobID": "1220edd7-6904-44c3-bb35-5679fe3dee74", "numQueues": 5, "numMappers": 1}'


clean-iam:
	awslocal iam detach-role-policy --role-name ribble --policy-arn arn:aws:iam::000000000000:policy/ribblePolicy
	awslocal iam detach-user-policy --user-name jose --policy-arn arn:aws:iam::000000000000:policy/assumeRibblePolicy
	awslocal iam delete-policy --policy-arn arn:aws:iam::000000000000:policy/ribblePolicy
	awslocal iam delete-policy --policy-arn arn:aws:iam::000000000000:policy/assumeRibblePolicy
	awslocal iam delete-role --role-name ribble

list-iam:
	awslocal iam list-policies | grep ribble
	awslocal iam list-policies | grep Ribble
	awslocal iam list-roles | grep ribble

test:
	echo "hello Ribble"


build-integration:
	docker build -f ./build/integration_tests/Dockerfile -t integration:latest .

install-aws:
	curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
	# unzip ./awscliv2.zip
	# sudo ./aws/install
	# aws --version

integration-s3:
	awslocal s3 mb s3://integration-test-bucket
	awslocal s3 cp ./build/integration_tests/test_data/test_lineitem.tbl.1  s3://integration-test-bucket/test_lineitem.tbl.1
	awslocal s3 cp ./build/integration_tests/test_data/test_lineitem.tbl.2  s3://integration-test-bucket/test_lineitem.tbl.2