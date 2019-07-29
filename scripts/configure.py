#!/usr/bin/env python

import os
import argparse
import json
from functools import reduce


def append_codedir(files):
    return map(lambda x: '$(CODE_DIR)/' + x, files)


def get_common_sources(workdir):
    dirs = [workdir, os.path.join(workdir, "functions")]
    files = [os.path.join(d, f) for d in dirs for f in os.listdir(d)]
    codes = filter(lambda x: x.endswith(".go")
                   and not x.endswith("_test.go"), files)
    return append_codedir(codes)


def get_functions():
    basedir = './functions'
    return filter(lambda x: os.path.isdir(os.path.join(basedir, x)),
                  os.listdir(basedir))


def get_test_functions():
    basedir = './test'
    return filter(lambda x: os.path.isdir(os.path.join(basedir, x)),
                  os.listdir(basedir))


def generate_header():
    required_parameters = [
        'StackName',
        'TestStackName',
        'Region',
        'CodeS3Bucket',
        'CodeS3Prefix',
    ]

    functions = map(lambda f: os.path.join('build', f), get_functions())
    test_functions = map(lambda f: os.path.join(
        'build', f), get_test_functions())

    lines = [
        'DEPLOY_CONFIG ?= deploy.jsonnet',
        'STACK_CONFIG ?= stack.jsonnet',
        'TEST_STACK_CONFIG ?= test.jsonnet'
        '',
        'CODE_DIR := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))',
        'CWD := ${CURDIR}',
        '',
        'TEMPLATE_JSONNET=$(CODE_DIR)/template.libsonnet',
        'TEMPLATE_FILE=template.json',
        'SAM_FILE=sam.yml',
        'OUTPUT_FILE=$(CWD)/output.json',
        '',
        'TEST_TEMPLATE_JSONNET=teststack.libsonnet',
        'TEST_TEMPLATE_FILE=teststack_template.json',
        'TEST_SAM_FILE=teststack_sam.yml',
        'TEST_OUTPUT_FILE=$(CWD)/teststack_output.json',
        '',
        'COMMON={}'.format(' '.join(get_common_sources('.'))),
        'FUNCTIONS={}'.format(' '.join(functions)),
        'TEST_FUNCTIONS={}'.format(' '.join(test_functions)),
        'TEST_UTILS=$(CODE_DIR)/test/*.go',
        '',
    ] + [
        '{0}=$(shell jsonnet $(DEPLOY_CONFIG) | jq .{0})'.format(p) for p in required_parameters
    ]

    return '\n'.join(lines)


def generate_functions_section():
    template = '''build/{0}: $(CODE_DIR)/functions/{0}/*.go $(COMMON)
	cd $(CODE_DIR) && env GOARCH=amd64 GOOS=linux go build -o $(CWD)/build/{0} ./functions/{0} && cd $(CWD)'''

    header = ['', '# Functions ------------------------']
    lines = map(lambda x: template.format(x), get_functions())
    return '\n'.join(header + lines) + '\n'


def generate_test_functions_section():
    template = '''build/{0}: $(CODE_DIR)/test/{0}/*.go
	cd $(CODE_DIR) && env GOARCH=amd64 GOOS=linux go build -o $(CWD)/build/{0} ./test/{0} && cd $(CWD)'''

    header = ['', '# TestStack Functions ------------------------']
    test_function_builds = [template.format(x) for x in get_test_functions()]

    return '\n'.join(header + test_function_builds + [''])


def get_all_source_files(workdir):
    return [os.path.join(dpath, fname) for dpath, dnames,
            files in os.walk(workdir) for fname in files
            if fname.endswith('.go')]


def generate_task_section():
    return '''
# Base Tasks -------------------------------------
build: $(FUNCTIONS)

clean:
	rm $(FUNCTIONS)

$(TEMPLATE_FILE): $(STACK_CONFIG) $(TEMPLATE_JSONNET)
	jsonnet -J $(CODE_DIR) $(STACK_CONFIG) -o $(TEMPLATE_FILE)

$(SAM_FILE): $(TEMPLATE_FILE) $(FUNCTIONS)
	aws cloudformation package \\
		--template-file $(TEMPLATE_FILE) \\
		--s3-bucket $(CodeS3Bucket) \\
		--s3-prefix $(CodeS3Prefix) \\
		--output-template-file $(SAM_FILE)

$(OUTPUT_FILE): $(SAM_FILE)
	aws cloudformation deploy \\
		--region $(Region) \\
		--template-file $(SAM_FILE) \\
		--stack-name $(StackName) \\
		--no-fail-on-empty-changeset \\
		--capabilities CAPABILITY_IAM $(PARAMETERS)
	aws cloudformation describe-stack-resources --stack-name $(StackName) > $(OUTPUT_FILE)

deploy: $(OUTPUT_FILE)

test: $(TEST_OUTPUT_FILE) $(TEST_UTILS)
	cd $(CODE_DIR) && env DEEPALERT_STACK_OUTPUT=$(OUTPUT_FILE) DEEPALERT_TEST_STACK_OUTPUT=$(TEST_OUTPUT_FILE) go test -v -count=1 . ./functions && cd $(CWD)

$(TEST_TEMPLATE_FILE): $(TEST_STACK_CONFIG) $(TEMPLATE_JSONNET)
	jsonnet -J $(CODE_DIR) $(TEST_STACK_CONFIG) -o $(TEST_TEMPLATE_FILE)

$(TEST_SAM_FILE): $(TEST_TEMPLATE_FILE) $(TEST_FUNCTIONS) $(OUTPUT_FILE)
	aws cloudformation package \\
		--template-file $(TEST_TEMPLATE_FILE) \\
		--s3-bucket $(CodeS3Bucket) \\
		--s3-prefix $(CodeS3Prefix) \\
		--output-template-file $(TEST_SAM_FILE)

$(TEST_OUTPUT_FILE): $(TEST_SAM_FILE)
	aws cloudformation deploy \\
		--region $(Region) \\
		--template-file $(TEST_SAM_FILE) \\
		--stack-name $(TestStackName) \\
		--capabilities CAPABILITY_IAM \\
		--no-fail-on-empty-changeset
	aws cloudformation describe-stack-resources --region $(Region) \\
        --stack-name $(TestStackName) > $(TEST_OUTPUT_FILE)

setuptest: $(TEST_OUTPUT_FILE)
'''


def main():
    psr = argparse.ArgumentParser()
    psr.add_argument('-o', '--output', default="Makefile")

    args = psr.parse_args()

    body = [
        generate_header(),
        generate_functions_section(),
        generate_task_section(),
        generate_test_functions_section(),
    ]

    makefile = '\n'.join(body)
    if args.output != "-":
        open(args.output, 'wt').write(makefile)
    else:
        print(makefile)


if __name__ == '__main__':
    main()
