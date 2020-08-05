#!/usr/bin/env node
import 'source-map-support/register';
import * as cdk from '@aws-cdk/core';
import { DeepalertStack } from '../lib/deepalert-stack';

const app = new cdk.App();
new DeepalertStack(app, 'DeepalertStack');
