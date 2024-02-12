package main

import (
	"context"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/haandol/vertical-slice-go-lambda-example/api/internal/feature/clickstream/handler"
	"github.com/haandol/vertical-slice-go-lambda-example/api/pkg/middleware"
	"github.com/haandol/vertical-slice-go-lambda-example/api/pkg/o11y"
	"github.com/haandol/vertical-slice-go-lambda-example/api/pkg/util/slogger"
)

const (
	isProd bool = true
)

var (
	r         *gin.Engine
	ginLambda *ginadapter.GinLambdaV2
)

func init() {
	// setup logger
	logger := slogger.Init(isProd)
	logger.Info("initializing...")

	// setup o11y
	o11y.InitXray(logger)

	// setup router
	r = gin.Default()
	r.Use(middleware.GinXrayMiddleware("ClickStreamService"))
	r.Use(middleware.GinSlogWithConfig(logger, &middleware.Config{
		UTC: false,
	}))
	r.Use(cors.New(cors.Config{
		AllowOrigins:  []string{"*"},
		AllowMethods:  []string{"*"},
		AllowHeaders:  []string{"*"},
		ExposeHeaders: []string{"Content-Length"},
		MaxAge:        12 * time.Hour,
	}))
	r.Use(middleware.RecoveryWithSlog(logger, true))

	rg := r.Group("/v1/clickstream")
	rg.POST("/:path", handler.CreateClickEventController)
	rg.GET("/:path", handler.GetClickStreamController)

	// setup ginLambda
	ginLambda = ginadapter.NewV2(r)
}

func LambdaHandler(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	if req.RequestContext.HTTP.Method == http.MethodOptions {
		return events.APIGatewayV2HTTPResponse{
			StatusCode: http.StatusOK,
			Headers: map[string]string{
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Methods": "POST, GET, PUT, DELETE, OPTIONS",
				"Access-Control-Allow-Headers": "Content-Type, Authorization, X-Amzn-Trace-Id, X-Requested-With",
				"Access-Control-Max-Age":       "3600",
			},
		}, nil
	}

	return ginLambda.ProxyWithContext(ctx, req)
}

func main() {
	lambda.Start(LambdaHandler)
}
