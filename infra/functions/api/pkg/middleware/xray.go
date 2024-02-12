package middleware

import (
	"net/http"
	"strings"

	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/gin-gonic/gin"
)

func GinXrayMiddleware(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, seg := xray.BeginSegmentWithSampling(c.Request.Context(), serviceName, c.Request, nil)
		defer seg.Close(nil)

		c.Request = c.Request.WithContext(ctx)

		captureRequestData(c, seg)
		c.Next()
		captureResponseData(c, seg)
	}
}

// Write request data to segment
func captureRequestData(c *gin.Context, seg *xray.Segment) {
	req := c.Request

	seg.Lock()
	defer seg.Unlock()
	segmentRequest := seg.GetHTTP().GetRequest()
	segmentRequest.Method = req.Method
	segmentRequest.URL = req.URL.String()
	segmentRequest.XForwardedFor = hasXForwardedFor(req)
	segmentRequest.ClientIP = clientIP(req)
	segmentRequest.UserAgent = req.UserAgent()
}

// Write response data to segment
func captureResponseData(c *gin.Context, seg *xray.Segment) {
	respStatus := c.Writer.Status()

	seg.Lock()
	defer seg.Unlock()
	segmentResponse := seg.GetHTTP().GetResponse()
	segmentResponse.Status = respStatus
	segmentResponse.ContentLength = c.Writer.Size()

	if respStatus >= 400 && respStatus < 500 {
		seg.Error = true
	}
	if respStatus == 429 {
		seg.Throttle = true
	}
	if respStatus >= 500 && respStatus < 600 {
		seg.Fault = true
	}
}

func hasXForwardedFor(r *http.Request) bool {
	return r.Header.Get("X-Forwarded-For") != ""
}

func clientIP(r *http.Request) string {
	forwardedFor := r.Header.Get("X-Forwarded-For")
	if forwardedFor != "" {
		return strings.TrimSpace(strings.Split(forwardedFor, ",")[0])
	}

	return r.RemoteAddr
}
