# displ

## Build image
docker build -t mapper .

## Test locally 
mkdir -p ~/.aws-lambda-rie && curl -Lo ~/.aws-lambda-rie/aws-lambda-rie \
https://github.com/aws/aws-lambda-runtime-interface-emulator/releases/latest/download/aws-lambda-rie-arm64  \
&& chmod +x ~/.aws-lambda-rie/aws-lambda-rie    

docker run -d -v ~/.aws-lambda-rie:/aws-lambda --entrypoint /aws-lambda/aws-lambda-rie -p 9000:8080 mapper /mapper     

curl -XPOST "http://localhost:9000/2015-03-31/functions/function/invocations" -d '{}'
	"bucket": "jobbucket",
	"id": "fa44fa0d-89de-416f-badc-00e2300acad2"
	"queues": 10,
	"size": 7375844,
	"rangeObjects": [{
        "objectBucket":"mybucket",
        "objectKey":"file1.csv",
		"initialByte": 0,
		"finalByte": 1494763,
    }, {
        "objectBucket":"mysecondbucket",
        "objectKey":"file3.csv",
		"initialByte": 0,
		"finalByte": 5881081,
    }]
}'

