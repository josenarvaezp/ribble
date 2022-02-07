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
	curl -XPOST "http://localhost:9000/2015-03-31/functions/function/invocations" -d '{"jobID": "ea11106c-27cd-45d8-8fae-2315d3c62248", "mapping": {"id": "fa44fa0d-89de-416f-badc-00e2300acad2", "queues": 5, "size": 2713576,"rangeObjects": [{"objectBucket":"mybucket", "objectKey":"file1.csv","initialByte": 0,"finalByte": 1356788}, {"objectBucket":"mysecondbucket","objectKey":"file3.csv","initialByte": 0,"finalByte": 1356788}]}}'

runr:
	curl -XPOST "http://localhost:9001/2015-03-31/functions/function/invocations" -d '{"jobID": "c7559a18-deda-44be-ad62-e33cd1546d09", "jobBucket": "jobbucket", "queueName": "e9a2c4c7-abb3-4c19-87bc-62e221da864a"}'

runc:
	curl -XPOST "http://localhost:9002/2015-03-31/functions/function/invocations" -d '{"jobID": "39ca4104-3278-4c95-a993-cc183b55485b", "numQueues": 5, "numMappers": 1}'

buildr:
	docker build -t reducer .

geto:
	awslocal s3api get-object --bucket e33e7151-8ae8-4560-93b2-5ba389949f7d --key output/7ecd2711-56d1-4d01-90a2-90b80e92c7e1 res
