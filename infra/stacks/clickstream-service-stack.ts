import * as path from 'path';
import * as cdk from 'aws-cdk-lib';
import { Construct } from 'constructs';
import * as iam from 'aws-cdk-lib/aws-iam';
import * as lambda from 'aws-cdk-lib/aws-lambda';
import * as lambdaGo from '@aws-cdk/aws-lambda-go-alpha';
import * as apigw from 'aws-cdk-lib/aws-apigatewayv2';
import * as cloudwatch from 'aws-cdk-lib/aws-cloudwatch';
import * as dynamodb from 'aws-cdk-lib/aws-dynamodb';
import { HttpLambdaIntegration } from 'aws-cdk-lib/aws-apigatewayv2-integrations';

interface IProps extends cdk.StackProps {
  api: apigw.IHttpApi;
  tableName: string;
}

export class ClickstreamServiceStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props: IProps) {
    super(scope, id, props);

    this.newClickstreamTable(props);

    const fn = this.newClickstreamServiceFunction(props);
    this.registerClickstreamServiceRoute(fn, props.api);
  }

  private newClickstreamTable(props: IProps) {
    return new dynamodb.Table(this, 'ClickstreamTable', {
      tableName: props.tableName,
      partitionKey: { name: 'PK', type: dynamodb.AttributeType.STRING },
      sortKey: { name: 'SK', type: dynamodb.AttributeType.STRING },
      encryption: dynamodb.TableEncryption.AWS_MANAGED,
      billingMode: dynamodb.BillingMode.PAY_PER_REQUEST,
    });
  }

  private newClickstreamServiceFunction(props: IProps) {
    const ns = this.node.tryGetContext('ns') as string;

    const fn = new lambdaGo.GoFunction(this, 'ClickstreamService', {
      functionName: `${ns}ClickstreamService`,
      entry: path.resolve(
        __dirname,
        '..',
        'functions',
        'api',
        'cmd',
        'clickstream'
      ),
      runtime: lambda.Runtime.PROVIDED_AL2023,
      architecture: lambda.Architecture.ARM_64,
      timeout: cdk.Duration.seconds(5),
      bundling: {
        goBuildFlags: ['-ldflags "-s -w"'],
      },
      environment: {
        AWS_XRAY_TRACING_NAME: 'ClickStreamService',
      },
    });
    fn.addToRolePolicy(
      new iam.PolicyStatement({
        actions: [
          'dynamodb:Query',
          'dynamodb:Scan',
          'dynamodb:GetItem',
          'dynamodb:PutItem',
          'dynamodb:UpdateItem',
          'dynamodb:DeleteItem',
        ],
        resources: [
          `arn:aws:dynamodb:${this.region}:${this.account}:table/${props.tableName}`,
          `arn:aws:dynamodb:${this.region}:${this.account}:table/${props.tableName}/index/*`,
        ],
      })
    );
    fn.addToRolePolicy(
      new iam.PolicyStatement({
        actions: ['xray:*'],
        resources: ['*'],
      })
    );
    cloudwatch.Metric.grantPutMetricData(fn);
    return fn;
  }

  private registerClickstreamServiceRoute(
    fn: lambda.IFunction,
    httpApi: apigw.IHttpApi
  ) {
    const integration = new HttpLambdaIntegration(
      'ClickstreamServiceInteg',
      fn
    );
    new apigw.HttpRoute(this, 'CorsClickstreamServiceRouteV1', {
      httpApi,
      routeKey: apigw.HttpRouteKey.with(
        '/v1/clickstream/{proxy+}',
        apigw.HttpMethod.OPTIONS
      ),
      integration,
      authorizer: new apigw.HttpNoneAuthorizer(),
    });
    new apigw.HttpRoute(this, 'ClickstreamServiceRouteV1', {
      httpApi,
      routeKey: apigw.HttpRouteKey.with(
        '/v1/clickstream/{proxy+}',
        apigw.HttpMethod.ANY
      ),
      integration,
    });
  }
}
