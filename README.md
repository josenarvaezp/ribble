# Ribble

## Prerequisites
To run Ribble you need to have the following in your local machine:

- Docker 
- AWS CLI installed and configured. Instructions can be found [here](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html)
- To setup ribble your AWS user needs to have AdministratorAccess or permission to create roles and policies
- Go (at least version 1.16)
- Make

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

The `build` command is used to create the resources locally that are needed to run the job. Specifically it auto-generates AWS lambda complacent `.go` files for the job coordinator and the mapper function. It also auto-generates `Dockerfiles` for each of them and builds the images.

An example of how to define a job can be found here: [Word count example](https://github.com/josenarvaezp/ribble/tree/main/examples/wordcount). It is simple, you need to define your map function and then create a main package where you need to define the ribble job using the function `ribble.Job()` from the [ribble package](https://github.com/josenarvaezp/ribble/tree/main/pkg/ribble/ribble.go). 

The map function you define has some restrictions:
1. It must take a string as input and this string is the file name of a file to process
2. It must return one of the aggregators defined here: [ribble aggregators](https://github.com/josenarvaezp/ribble/tree/main/pkg/aggregators/aggregators.go)
The available aggregators include:
- `MapSum`: MapSum adds all values that have the same key
- `MapMax`: MapMax gets the maximum value in each key
- `MapMin`: MapMin gets the minimum value in each key

```
ribble build --job <path-to-your-job-definition>
```

## Upload

The `upload` command is used to upload all the resources that were genererated by the `build` command and creates additional resources needed to run the ribble job such as the SQS queues that hold the intermediate data, a bucket for the job, amongst other. 

```
ribble upload --job-id <id-of-job>
```

## Run

The `run` command is used to run the job with the given job id. Note that this command runs the ribble job but does not wait until it has completed. If any errors occurred or you want to know the status of the job you need to use `TODO`

```
ribble upload --job-id <id-of-job>
```

# Local testing

For local testing you can use Localstack, a docker service that replicates AWS locally. You can either use the AWS CLI by using the `--endpoint-url` flag like: `aws --endpoint-url=http://localhost:4566 s3 ls` or you can download [awslocal](https://github.com/localstack/awscli-local).

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
