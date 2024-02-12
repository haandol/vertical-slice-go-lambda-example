package handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/gin-gonic/gin"
	"github.com/haandol/vertical-slice-go-lambda-example/api/internal/feature/clickstream/domain"
	"github.com/haandol/vertical-slice-go-lambda-example/api/internal/feature/clickstream/instrument"
	commoninstrument "github.com/haandol/vertical-slice-go-lambda-example/api/internal/instrument"
	"github.com/haandol/vertical-slice-go-lambda-example/api/pkg/connector/cloud"
	"github.com/haandol/vertical-slice-go-lambda-example/api/pkg/o11y"
	"github.com/haandol/vertical-slice-go-lambda-example/api/pkg/util/slogger"
)

// primary adapter.
func GetClickStreamController(c *gin.Context) {
	logger := slogger.New().WithArgs(
		"feature", "clickstream",
		"usecase", "getClickStream",
		"component", "controller",
	)

	ctx, cancel := context.WithTimeout(c.Request.Context(), ControllerTimeout)
	defer cancel()

	ctx, span := o11y.BeginSubSegment(ctx, "GetClickStreamController")
	defer span.Close(nil)

	path := c.Param("path")

	result, err := GetClickStreamService(ctx, path)
	if err != nil {
		commoninstrument.RecordError(logger, span, err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "failed to get clickstream for path",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": len(result),
	})
}

// service.
func GetClickStreamService(ctx context.Context, path string) ([]domain.ClickEvent, error) {
	logger := slogger.New().WithContext(ctx).WithArgs(
		"feature", "clickstream",
		"usecase", "getClickStream",
		"component", "service",
		"path", path,
	)

	ctx, span := o11y.BeginSubSegment(ctx, "GetClickStreamService")
	defer span.Close(nil)

	var clickstream []domain.ClickEvent

	clickstream, err := GetClickStreamRepository(ctx, path)
	if err != nil {
		instrument.RecordGetClickStreamError(logger, span, err)
		return clickstream, err
	}

	return clickstream, nil
}

// secondary adapter.
func GetClickStreamRepository(ctx context.Context, path string) ([]domain.ClickEvent, error) {
	logger := slogger.New().WithContext(ctx).WithArgs(
		"feature", "clickstream",
		"usecase", "getClickStream",
		"component", "repository",
		"path", path,
	)

	ctx, span := o11y.BeginSubSegment(ctx, "GetClickStreamRepository")
	defer span.Close(nil)

	var clickstream []domain.ClickEvent

	client, err := cloud.NewDynamoDBClient()
	if err != nil {
		commoninstrument.RecordError(logger, span, err)
		return clickstream, err
	}

	params := &dynamodb.QueryInput{
		TableName:              aws.String("clickstream"),
		KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :sk)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: fmt.Sprintf("CLICK#EVENT#PATH#%v", path)},
			":sk": &types.AttributeValueMemberS{Value: "CLICK#EVENT#"},
		},
		ProjectionExpression: aws.String("ID"),
		Limit:                aws.Int32(ClickEventCountLimit),
	}
	output, err := client.Query(ctx, params)

	if err != nil {
		commoninstrument.SetParamsToSpanAttr(logger, span, params)
		commoninstrument.RecordError(logger, span, err)
		return clickstream, err
	}

	if err := attributevalue.UnmarshalListOfMaps(output.Items, &clickstream); err != nil {
		commoninstrument.RecordError(logger, span, err)
		return clickstream, err
	}

	return clickstream, nil
}
