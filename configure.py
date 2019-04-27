#!/usr/bin/env python

import os
import argparse


def get_functions():
    basedir = './functions'
    return filter(lambda x: os.path.isdir(os.path.join(basedir, x)),
                  os.listdir(basedir))


def gen_functions_section():
    template = '''build/{0}: ./functions/{0}/*.go $(COMMON)
	env GOARCH=amd64 GOOS=linux go build -o build/{0} ./functions/{0}/'''

    lines = map(lambda x: template.format(x), get_functions())
    return '\n'.join(lines)


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

StackName    := $(shell cat $(DEEPALERT_CONFIG) | jq '.["StackName"]' -r)
Region       := $(shell cat $(DEEPALERT_CONFIG) | jq '.["Region"]' -r)
CodeS3Bucket := $(shell cat $(DEEPALERT_CONFIG) | jq '.["CodeS3Bucket"]' -r)
CodeS3Prefix := $(shell cat $(DEEPALERT_CONFIG) | jq '.["CodeS3Prefix"]' -r)

functions: $(FUNCTIONS)

clean:
	rm $(FUNCTIONS)

$(OUTPUT_FILE): $(TEMPLATE_FILE) $(FUNCTIONS)
	aws cloudformation package \
		--template-file $(TEMPLATE_FILE) \
		--s3-bucket $(CodeS3Bucket) \
		--s3-prefix $(CodeS3Prefix) \
		--output-template-file $(OUTPUT_FILE)

deploy: $(OUTPUT_FILE)
	aws cloudformation deploy \
		--region $(Region) \
		--template-file $(OUTPUT_FILE) \
		--stack-name $(StackName) \
		--capabilities CAPABILITY_IAM
'''
    functions = map(lambda f: os.path.join('build', f), get_functions())
    return template.format(' '.join(functions))


def main():
    psr = argparse.ArgumentParser()
    psr.add_argument('-o', '--output', default="Makefile")

    args = psr.parse_args()

    body = [
        gen_header(),
        gen_functions_section(),
    ]

    makefile = '\n'.join(body)
    if args.output != "-":
        open(args.output, 'wt').write(makefile)
    else:
        print(makefile)


if __name__ == '__main__':
    main()
