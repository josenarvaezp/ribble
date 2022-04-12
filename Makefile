build_cli:
	go build -o ./ribble ./cmd/driver/main.go

integration-s3:
	awslocal s3 mb s3://integration-test-bucket
	awslocal s3 cp ./build/integration_tests/test_data/test_lineitem.tbl.1  s3://integration-test-bucket/test_lineitem.tbl.1
	awslocal s3 cp ./build/integration_tests/test_data/test_lineitem.tbl.2  s3://integration-test-bucket/test_lineitem.tbl.2

remove_images:
	docker images | grep word | awk '{ print $3; }' | xargs docker rmi

download-output:
	awslocal s3api get-object --bucket 2f18f4e1-60e7-491f-97b7-989c996f577e --key output oooout4

create-mocks:
	mockery --dir ./internal --all 