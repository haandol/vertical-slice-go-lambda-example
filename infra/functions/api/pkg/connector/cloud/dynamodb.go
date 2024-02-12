package cloud

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/haandol/vertical-slice-go-lambda-example/api/pkg/util"
	"github.com/haandol/vertical-slice-go-lambda-example/api/pkg/util/slogger"
)

var (
	ddbClient          *dynamodb.Client
	LOCAL_DDB_ENDPOINT = os.Getenv("LOCAL_DDB_ENDPOINT")
)

func NewDynamoDBClient() (*dynamodb.Client, error) {
	logger := slogger.New()

	if ddbClient != nil {
		logger.Info("client already initialized")
		return ddbClient, nil
	}

	if util.IsLocalEnv() && LOCAL_DDB_ENDPOINT != "" {
		logger.Warn("using local dynamodb", "endpoint", LOCAL_DDB_ENDPOINT)
		ddbClient = newLocalClient()
		return ddbClient, nil
	}

	awsCfg, err := GetAWSConfig()
	if err != nil {
		logger.Error("failed to get aws config", "err", err)
		return nil, err
	}

	ddbClient = dynamodb.NewFromConfig(*awsCfg)
	return ddbClient, nil
}

func newLocalClient() *dynamodb.Client {
	awsCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
			func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{URL: LOCAL_DDB_ENDPOINT}, nil
			})),
	)
	if err != nil {
		panic(err)
	}

	return dynamodb.NewFromConfig(awsCfg)
}
