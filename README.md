# Ribble

## Prerequisites
To run Ribble you need to have the following in your local machine:

- Docker 
- AWS CLI installed and configured. Instructions can be found at https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html
- To setup ribble your AWS user needs to have AdministratorAccess or permission to create roles and policies
- Go (at least version 1.16)
- Make

## Download the code
```
git clone git@github.com:josenarvaezp/ribble.git
```

## build Ribble CLI
To build  the CLI tool run:
```
make build_cli
```

## Set credentials

Ribble needs AWS permissions to access S3, SQS, Lambda, IAM, and ECR to run a processing job. To facilitate adding these permissions into your account you can use the `set-credentials` command.  The `set-credentials` command creates an AWS role called `ribble` and it assigns to it the policies it requires to access the resources needed. It then gives the specified user permission to assume this role. Hence this command can be used by an administrator in the AWS account (someone with AWS AdministratorAccess) to give access to different users within the account that need to run ribble jobs. 

```
ribble set-credentials \
    --account-id <your-account-id> \
    --username <aws-username>
```

Use --local to create the credentials in localstack.

## Build

The `build` command is used to create the resources locally that are needed to run the job. Specifically it auto-generates AWS lambda `.go` files for the job coordinator, mapper and reducer functions. It also auto-generates `Dockerfiles` for each of them and builds the images.

```
ribble build --job <path-to-your-job-definition>
```

Output:
```
Generating resources...
Building docker images...
Build successful with Job ID:  308866c6-2ef0-4f80-868e-6b1760da8eb9
```

## Upload

The `upload` command is used to upload all the resources that were genererated by the `build` command and creates additional resources needed to run the ribble job such as the SQS queues that hold the intermediate data, a bucket for the job, amongst other. 

```
ribble upload --job-id <id-of-job>
```

Output:
```
Creating resources...
Creating job S3 bucket...
Generating mappings...
Writing mappings to S3...
Creating streams in SQS...
Creating log stream in CloudWatch...
Creating SQS dead-letter queue...
Uploading Lambda functions...
Upload successful with Job ID:  308866c6-2ef0-4f80-868e-6b1760da8eb9
```

## Run

The `run` command is used to run the job with the given job id. Note that this command runs the ribble job but does not wait until it has completed. If any errors occurred or you want to know the status of the job you need to use the `track` command.

```
ribble run --job-id <id-of-job>
```

Output:
```
Running job:  308866c6-2ef0-4f80-868e-6b1760da8eb9
```

## Track

The `track` command is used to track the progress of a job. It can tell you how many mappers and reducers are left in the job or if the job has been completed. 

```
ribble track --job-id <id-of-job>
```

Output:
```
INFO[0000] Coordinator starting...                       Timestamp="54305-01-28 13:11:09 +0000 GMT"
INFO[0000] Waiting for 1 mappers...                      Timestamp="54305-01-28 13:11:09 +0000 GMT"
INFO[0000] Mappers execution completed...                Timestamp="54305-01-28 14:06:08 +0000 GMT"
INFO[0000] Waiting for 1 reducers...                     Timestamp="54305-01-28 14:06:08 +0000 GMT"
INFO[0000] Reducers execution completed...               Timestamp="54305-01-28 14:58:52 +0000 GMT"
INFO[0000] Waiting for final reducer...                  Timestamp="54305-01-28 14:58:52 +0000 GMT"
INFO[0000] Final reducer execution completed...          Timestamp="54305-01-28 16:23:16 +0000 GMT"
INFO[0000] Job completed successfully, output is available at the S3 bucket 308866c6-2ef0-4f80-868e-6b1760da8eb9...  Timestamp="54305-01-28 16:23:16 +0000 GMT"
```

## Local testing

For local testing you can use Localstack, a docker service that replicates AWS locally. You can either use the AWS CLI by using the `--endpoint-url` flag like: `aws --endpoint-url=http://localhost:4566 s3 ls` or you can download awslocal at https://github.com/localstack/awscli-local.

To start localstack run:
```
docker-compose up -d
```

Once localstack is running, create a bucket and upload files:
```
awslocal s3 mb s3://my-input-bucket
awslocal s3 cp test.txt  s3://my-input-bucket/test.txt
```

When setting up the job, remember to set `Local` to true in the configuration. You can then use the ribble CLI as described in the first section. While you can test the `set-credentials` command by setting the `--local` flag to true, note that localstack does not check IAM roles or users so you can skip that.

Once the jobs finishes you will be able to get the result from S3 using the awslocal CLI.
