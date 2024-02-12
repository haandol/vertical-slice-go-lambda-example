#!/usr/bin/env node
import 'source-map-support/register';
import * as cdk from 'aws-cdk-lib';
import { GatewayStack } from '../stacks/gateway-stack';
import { ClickstreamServiceStack } from '../stacks/clickstream-service-stack';
import { Config } from '../config/loader';

const app = new cdk.App({
  context: {
    ns: Config.app.ns,
    stage: Config.app.stage,
  },
});

const gatewayStack = new GatewayStack(app, `${Config.app.ns}GatewayStack`, {
  env: {
    region: Config.aws.region,
  },
});

const clickStreamService = new ClickstreamServiceStack(
  app,
  `${Config.app.ns}ClickStreamServiceStack`,
  {
    api: gatewayStack.api,
    tableName: Config.table.clickstream.name,
    env: {
      region: Config.aws.region,
    },
  }
);
clickStreamService.addDependency(gatewayStack);
