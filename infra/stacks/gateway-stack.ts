import * as path from 'path';
import * as cdk from 'aws-cdk-lib';
import * as apigw from 'aws-cdk-lib/aws-apigatewayv2';
import * as authorizers from 'aws-cdk-lib/aws-apigatewayv2-authorizers';
import * as lambda from 'aws-cdk-lib/aws-lambda';
import * as lambdaNodejs from 'aws-cdk-lib/aws-lambda-nodejs';
import { Construct } from 'constructs';

interface IProps extends cdk.StackProps {
  authToken: string;
}

export class GatewayStack extends cdk.Stack {
  public readonly api: apigw.IHttpApi;

  constructor(scope: Construct, id: string, props: IProps) {
    super(scope, id, props);

    const ns = this.node.tryGetContext('ns') as string;

    const authorizer = this.newLambdaAuthorizer(props);
    this.api = this.newHttpApi(ns, authorizer);
  }

  private newLambdaAuthorizer(props: IProps): authorizers.HttpLambdaAuthorizer {
    const fn = new lambdaNodejs.NodejsFunction(this, 'AuthorizerFunction', {
      entry: path.resolve(
        __dirname,
        '..',
        'functions',
        'auth',
        'authorizer.ts'
      ),
      runtime: lambda.Runtime.NODEJS_20_X,
      architecture: lambda.Architecture.ARM_64,
      environment: {
        AUTH_TOKEN: props.authToken,
      },
    });
    return new authorizers.HttpLambdaAuthorizer('Authorizer', fn, {
      responseTypes: [authorizers.HttpLambdaResponseType.SIMPLE],
    });
  }

  private newHttpApi(
    ns: string,
    defaultAuthorizer: apigw.IHttpRouteAuthorizer
  ): apigw.HttpApi {
    const api = new apigw.HttpApi(this, 'HttpApi', {
      apiName: `${ns}Api`,
      corsPreflight: {
        allowOrigins: ['http://localhost:3000'],
        allowMethods: [
          apigw.CorsHttpMethod.POST,
          apigw.CorsHttpMethod.GET,
          apigw.CorsHttpMethod.PUT,
          apigw.CorsHttpMethod.DELETE,
          apigw.CorsHttpMethod.OPTIONS,
        ],
        allowHeaders: [
          'Authorization',
          'Content-Type',
          'X-Amzn-Trace-Id',
          'X-Requested-With',
        ],
        allowCredentials: false,
        maxAge: cdk.Duration.days(1),
      },
      defaultAuthorizer,
    });
    new cdk.CfnOutput(this, 'HttpApiUrl', {
      value: api.apiEndpoint,
    });
    return api;
  }
}
