# DeepAlert


## Prerequisite (deployment and test)

- Tools
  - `python` >= 3.6.4
  - `awscli` >= 1.16.140
  - `go` >= 1.12.4
  - `GNU Make` >= 3.81
- Credential
  - AWS CLI credential to deploy CloudFormation. See [here](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html) for more detail.

## How to use

### Getting Started

```shell
$ ./configure.py --StackName=your-deepalert-stack --CodeS3Bucket=YOUR_BUCKET_TO_STORE_BINARY --Region=ap-northeast-1
$ make deploy
```

### configure script

Options of `configure.py` to generate `Makefile` is following.

- CLI options
  - `-o, --output`: Specify output file name. `-` means stdout. Default is `Makefile`
  - `-c, --config`: Specify JSON format config file. A config file can have parameter values, e.g. `{"StackName": "your-stack-name"}`
  - `-w, --workdir`: Specify working directory for deploy and test. Default is `.`.
- Parameters
  - `--StackName`: Required. Specify stack name of DeepAlert deployed by CloudFormation.
  - `--Region`: Required. Specify AWS region such as `ap-northeast-1` to deploy.
  - `--CodeS3Bucket`: Required. Specify S3 bucket name to store executable binary of DeepAlert.
  - `--CodeS3Prefix`: Required. Specify S3 key prefix to store executable binary of DeepAlert. NOTE: `/` at tail of prefix is not needed.
  - `--LambdaRoleArn`: Optional. Specify IAM Role ARN for Lambda functions of DeepAlert. If not specified, CloudFormation will create own IAM Role for Lambda as resource `LambdaRole`
  - `--StepFunctionRoleArn`: Optional. Specify IAM Role ARN for Lambda functions of DeepAlert. If not specified, CloudFormation will create own IAM Role for Lambda as resource `StepFunctionRole`
  - `--ReviewerLambdaArn`: Optional. Specify "Reviewer" Lambda ARN. If not specified, `DummyReviewer` that evaluates all alert as `Unclassified` will be set.
  - `--InspectionDelay`: Optional. Specify delay seconds of invoking Inspectors. Default is `300` seconds.
  - `--ReviewDelay`: Optional. Specify delay seconds of invoking Reviewer. Default is `600` seconds.

CLI parameter (e.g. `--StackName`) overwrites same key parameter in config file specified by `--config` option.

## Development

### Architecture

![Archtecture](https://user-images.githubusercontent.com/605953/57503427-ff445600-732a-11e9-8089-953bc9cd0711.png)

### Test

```shell
$ ./configure.py --StackName=your-deepalert-test --CodeS3Bucket=YOUR_BUCKET_TO_STORE_BINARY --CodeS3Prefix=functions --Region=ap-northeast-1 -o Makefile.test --workdir=testing
$ make test -f Makefile.test
```

