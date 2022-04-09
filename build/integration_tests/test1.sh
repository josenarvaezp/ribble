#!/bin/sh

function fail() {
    echo $2
    exit $1
}

# Build job
RESPONSE=$(./ribble build --job ./build/integration_tests/tests/query1/job/query1_job.go | grep "Build successful with Job ID")
IFS=':'; JOB_ID=($RESPONSE); unset IFS;
JOB_ID=(${JOB_ID[1]})

[ $? == 0 ] || fail 1 "Failed building ribble job"

# upload job
./ribble upload --job-id ${JOB_ID}

[ $? == 0 ] || fail 1 "Failed uploading ribble job"

# run job
./ribble run --job-id ${JOB_ID}

[ $? == 0 ] || fail 1 "Failed running ribble job"

# verify job

awslocal s3api get-object --bucket ${JOB_ID} --key output/6391bbed-49fe-4d6c-b853-dd318a2ad99a q1_3gb_test
