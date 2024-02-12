package o11y

import (
	"context"
	"log"

	"github.com/aws/aws-xray-sdk-go/xray"
)

func InitXray() {
	xray.Configure(xray.Config{})
	log.Printf("X-Ray initialized")
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
