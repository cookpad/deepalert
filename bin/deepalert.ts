#!/usr/bin/env node
import "source-map-support/register";
import * as cdk from "@aws-cdk/core";
import { DeepAlertStack } from "../cdk/deepalert-stack";

const app = new cdk.App();
new DeepAlertStack(app, "DeepAlertStack", {});
