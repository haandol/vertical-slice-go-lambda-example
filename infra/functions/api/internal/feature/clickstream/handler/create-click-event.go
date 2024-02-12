package handler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/gin-gonic/gin"
	"github.com/haandol/vertical-slice-go-lambda-example/api/internal/feature/clickstream/domain"
	"github.com/haandol/vertical-slice-go-lambda-example/api/internal/feature/clickstream/dto"
	"github.com/haandol/vertical-slice-go-lambda-example/api/internal/feature/clickstream/instrument"
	commoninstrument "github.com/haandol/vertical-slice-go-lambda-example/api/internal/instrument"
	"github.com/haandol/vertical-slice-go-lambda-example/api/pkg/connector/cloud"
	"github.com/haandol/vertical-slice-go-lambda-example/api/pkg/o11y"
	"github.com/haandol/vertical-slice-go-lambda-example/api/pkg/util/slogger"
	"github.com/oklog/ulid/v2"
)

// primary adapter.
func CreateClickEventController(c *gin.Context) {
	logger := slogger.New().WithArgs(
		"feature", "clickstream",
		"usecase", "createClickEvent",
		"component", "controller",
	)

	ctx, cancel := context.WithTimeout(c.Request.Context(), ControllerTimeout)
	defer cancel()

	ctx, span := o11y.BeginSubSegment(ctx, "CreateClickEventController")
	defer span.Close(nil)

	req := &dto.ClickEvent{}
	if err := c.ShouldBindJSON(req); err != nil {
		commoninstrument.RecordBadInputError(logger, span, err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	event, err := CreateClickEventService(ctx, req)
	if err != nil {
		commoninstrument.RecordError(logger, span, err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"message": "failed to create click-event",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": event.DTO(),
	})
}

// service.
func CreateClickEventService(ctx context.Context, req *dto.ClickEvent) (domain.ClickEvent, error) {
	logger := slogger.New().WithContext(ctx).WithArgs(
		"feature", "clickstream",
		"usecase", "createClickEventService",
		"component", "service",
	)

	ctx, span := o11y.BeginSubSegment(ctx, "CreateClickEventService")
	defer span.Close(nil)
	commoninstrument.RecordRequest(logger, span, req)

	var event domain.ClickEvent

	req.ID = ulid.Make().String()
	req.CreatedAt = time.Now().UTC().Format(time.RFC3339)

	event, err := CreateClickEventRepository(ctx, req)
	if err != nil {
		instrument.RecordGetClickStreamError(logger, span, err)
		return event, err
	}

	instrument.RecordCreateClickEventSuccess(logger)

	return event, nil
}

// secondary adapter.
func CreateClickEventRepository(ctx context.Context, req *dto.ClickEvent) (domain.ClickEvent, error) {
	logger := slogger.New().WithContext(ctx).WithArgs(
		"feature", "clickstream",
		"usecase", "createClickEvent",
		"component", "repository",
	)

	ctx, span := o11y.BeginSubSegment(ctx, "CreateClickEventRepository")
	defer span.Close(nil)

	var event domain.ClickEvent

	client, err := cloud.NewDynamoDBClient()
	if err != nil {
		commoninstrument.RecordError(logger, span, err)
		return event, err
	}

	params := &dynamodb.PutItemInput{
		TableName: aws.String("clickstream"),
		Item: map[string]types.AttributeValue{
			"PK":        &types.AttributeValueMemberS{Value: fmt.Sprintf("CLICK#EVENT#PATH#%v", req.Path)},
			"SK":        &types.AttributeValueMemberS{Value: fmt.Sprintf("CLICK#EVENT#%v", req.ID)},
			"id":        &types.AttributeValueMemberS{Value: req.ID},
			"path":      &types.AttributeValueMemberS{Value: req.Path},
			"createdAt": &types.AttributeValueMemberS{Value: req.CreatedAt},
		},
		ConditionExpression: aws.String("attribute_not_exists(PK) AND attribute_not_exists(SK)"),
	}
	_, err = client.PutItem(ctx, params)
	if err != nil {
		commoninstrument.SetParamsToSpanAttr(logger, span, params)
		commoninstrument.RecordError(logger, span, err)
		return event, err
	}

	event.ID = req.ID
	event.Path = req.Path
	event.CreatedAt = req.CreatedAt

	return event, nil
}
