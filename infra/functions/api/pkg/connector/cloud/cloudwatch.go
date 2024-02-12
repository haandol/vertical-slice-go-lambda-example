package cloud

import (
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/haandol/vertical-slice-go-lambda-example/api/pkg/util/slogger"
)

var cwClient *cloudwatch.Client

func NewCloudwatchClient() (*cloudwatch.Client, error) {
	logger := slogger.New()

	if cwClient != nil {
		logger.Info("client already initialized")
		return cwClient, nil
	}

	awsCfg, err := GetAWSConfig()
	if err != nil {
		logger.Error("failed to get aws config", "err", err)
		return nil, err
	}

	cwClient = cloudwatch.NewFromConfig(*awsCfg)
	return cwClient, nil
}
