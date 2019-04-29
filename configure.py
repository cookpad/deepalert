#!/usr/bin/env python

import os
import argparse
import json


def get_functions():
    basedir = './functions'
    return filter(lambda x: os.path.isdir(os.path.join(basedir, x)),
                  os.listdir(basedir))


def gen_functions_section():
    template = '''build/{0}: ./functions/{0}/*.go $(COMMON)
	env GOARCH=amd64 GOOS=linux go build -o build/{0} ./functions/{0}/'''

    lines = map(lambda x: template.format(x), get_functions())
    return '\n'.join(lines) + '\n'


def gen_header():
    template = '''
DEEPALERT_CONFIG ?= config.json

ifeq (,$(wildcard $(DEEPALERT_CONFIG)))
    $(error $(DEEPALERT_CONFIG) is not found)
endif

TEMPLATE_FILE=template.yml
OUTPUT_FILE=sam.yml
COMMON=functions/*.go
FUNCTIONS={0}
'''
    functions = map(lambda f: os.path.join('build', f), get_functions())
    return template.format(' '.join(functions))


def gen_parameters(args):
    required_parameters = [
        'StackName',
        'Region',
        'CodeS3Bucket',
        'CodeS3Prefix',
    ]
    optional_parameters = [
        'LambdaRoleArn',
        'StepFunctionRoleArn',
        'ReviewerLambdaArn',
        'InspectionDelay',
        'ReviewDelay',
    ]
    params = required_parameters + optional_parameters

    if args.config:
        config = dict([(k, v) for k, v in json.load(args.config).items()])
    else:
        config = {}

    for k in params:
        if hasattr(args, k) and getattr(args, k) is not None:
            config[k] = getattr(args, k)

    lines = []

    # Check required parameters
    for k in required_parameters:
        if k not in config.keys():
            raise Exception(
                'Parameter "{0}" is required, but not found'.format(k))

        lines.append('{0}="{1}"'.format(k, config[k]))

    # Build override parameters
    options = []
    for k in optional_parameters:
        if k in config.keys():
            options.append('{0}="{1}"'.format(k, config[k]))

    opt = ' '.join(["--parameter-overrides"] + options) if options else ''

    lines.append('PARAMETERS={0}'.format(opt))

    return '\n'.join(lines) + '\n'


def gen_task_section():
    return '''
functions: $(FUNCTIONS)

clean:
	rm $(FUNCTIONS)

$(OUTPUT_FILE): $(TEMPLATE_FILE) $(FUNCTIONS)
	aws cloudformation package \\
		--template-file $(TEMPLATE_FILE) \\
		--s3-bucket $(CodeS3Bucket) \\
		--s3-prefix $(CodeS3Prefix) \\
		--output-template-file $(OUTPUT_FILE)

deploy: $(OUTPUT_FILE)
	aws cloudformation deploy \\
		--region $(Region) \\
		--template-file $(OUTPUT_FILE) \\
		--stack-name $(StackName) \\
		--capabilities CAPABILITY_IAM $(PARAMETERS)
'''


def main():
    psr = argparse.ArgumentParser()
    psr.add_argument('-o', '--output', default="Makefile")
    psr.add_argument('-c', '--config', type=argparse.FileType('rt'))
    psr.add_argument('--StackName')
    psr.add_argument('--Region')
    psr.add_argument('--CodeS3Bucket')
    psr.add_argument('--CodeS3Prefix')
    psr.add_argument('--LambdaRoleArn')
    psr.add_argument('--StepFunctionRoleArn')
    psr.add_argument('--ReviewerLambdaArn')
    psr.add_argument('--InspectionDelay', type=int)
    psr.add_argument('--ReviewDelay', type=int)

    args = psr.parse_args()

    body = [
        gen_header(),
        gen_parameters(args),
        gen_task_section(),
        gen_functions_section(),
    ]

    makefile = '\n'.join(body)
    if args.output != "-":
        open(args.output, 'wt').write(makefile)
    else:
        print(makefile)


if __name__ == '__main__':
    main()
