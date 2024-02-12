package cloud

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-xray-sdk-go/instrumentation/awsv2"
)

var (
	awsCfg      *aws.Config
	AWS_PROFILE = os.Getenv("AWS_PROFILE")
)

func GetAWSConfig() (*aws.Config, error) {
	if awsCfg != nil {
		return awsCfg, nil
	}

	optFns := []func(*config.LoadOptions) error{}
	if AWS_PROFILE != "" {
		fmt.Printf("use [%v] profile for AWS", AWS_PROFILE)
		optFns = append(optFns, config.WithSharedConfigProfile(AWS_PROFILE))
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(), optFns...)
	if err != nil {
		return awsCfg, err
	}
	awsCfg = &cfg
	awsv2.AWSV2Instrumentor(&cfg.APIOptions)

	return awsCfg, nil
}
