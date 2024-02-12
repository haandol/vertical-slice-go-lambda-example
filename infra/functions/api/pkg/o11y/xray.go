package o11y

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-xray-sdk-go/strategy/sampling"
	"github.com/aws/aws-xray-sdk-go/xray"
)

type Logger interface {
	Info(msg string, args ...any)
	Error(msg string, args ...any)
}

func InitXray(logger Logger) {
	rule := map[string]any{
		"version": 2,
		"rules": []map[string]any{
			{
				"description":  "All POST requests",
				"host":         "*",
				"http_method":  "POST",
				"url_path":     "*",
				"fixed_target": 0,
				"rate":         1.0,
			},
			{
				"description":  "All PUT requests",
				"host":         "*",
				"http_method":  "PUT",
				"url_path":     "*",
				"fixed_target": 0,
				"rate":         1.0,
			},
			{
				"description":  "All DELETE requests",
				"host":         "*",
				"http_method":  "DELETE",
				"url_path":     "*",
				"fixed_target": 0,
				"rate":         1.0,
			},
		},
		"default": map[string]any{
			"fixed_target": 0.0,
			"rate":         0.1,
		},
	}
	b, err := json.Marshal(rule)
	if err != nil {
		logger.Error("error on marshaling sampling strategy")
	}

	ss, err := sampling.NewCentralizedStrategyWithJSONBytes(b)
	if err != nil {
		logger.Error("error on new strategy from json")
	}

	if err := xray.Configure(xray.Config{
		SamplingStrategy: ss,
	}); err != nil {
		logger.Error("error on configure xray")
	}

	logger.Info("X-Ray initialized")
}

func BeginSegment(ctx context.Context, name string) (context.Context, *xray.Segment) {
	return xray.BeginSegment(ctx, name)
}

func BeginSubSegment(ctx context.Context, name string) (context.Context, *xray.Segment) {
	return xray.BeginSubsegment(ctx, name)
}

func GetTraceID(ctx context.Context) string {
	return xray.TraceID(ctx)
}
