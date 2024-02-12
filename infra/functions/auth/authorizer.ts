import { APIGatewayRequestAuthorizerEventV2 } from 'aws-lambda';

const AuthToken = process.env.AUTH_TOKEN || 'demo';

export const handler = async (event: APIGatewayRequestAuthorizerEventV2) => {
  console.log('Received event:', JSON.stringify(event));

  let response = {
    isAuthorized: false,
  };

  if (event.headers && event.headers.authorization === AuthToken) {
    console.log('allowed');
    response = {
      isAuthorized: true,
    };
  }

  return response;
};
