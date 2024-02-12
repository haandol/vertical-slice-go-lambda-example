package middleware

import (
	"bytes"
	"net/http"
	"strconv"
	"strings"

	"github.com/aws/aws-xray-sdk-go/header"
	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/gin-gonic/gin"
)

const (
	headerTraceID = "x-amzn-trace-id"
)

func GinXrayMiddleware(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		traceHeader := header.FromString(c.Request.Header.Get(headerTraceID))
		ctx, seg := xray.BeginSegmentWithSampling(c.Request.Context(), serviceName, c.Request, traceHeader)
		defer seg.Close(nil)

		c.Request = c.Request.WithContext(ctx)

		captureRequestData(c, seg)
		c.Next()
		captureResponseData(c, seg)
	}
}

// Write request data to segment.
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
	c.Writer.Header().Set(headerTraceID, createTraceHeader(req, seg))
}

// Write response data to segment.
func captureResponseData(c *gin.Context, seg *xray.Segment) {
	respStatus := c.Writer.Status()

	seg.Lock()
	defer seg.Unlock()
	segmentResponse := seg.GetHTTP().GetResponse()
	segmentResponse.Status = respStatus
	segmentResponse.ContentLength = c.Writer.Size()

	if respStatus >= http.StatusBadRequest && respStatus < http.StatusInternalServerError {
		seg.Error = true
	}
	if respStatus == http.StatusTooManyRequests {
		seg.Throttle = true
	}
	if respStatus >= http.StatusInternalServerError && respStatus < 600 {
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

func createTraceHeader(r *http.Request, seg *xray.Segment) string {
	trace := parseHeaders(r.Header)
	if trace["Root"] != "" {
		seg.TraceID = trace["Root"]
		seg.RequestWasTraced = true
	}
	if trace["Parent"] != "" {
		seg.ParentID = trace["Parent"]
	}
	// Don't use the segment's header here as we only want to
	// send back the root and possibly sampled values.
	var respHeader bytes.Buffer
	respHeader.WriteString("Root=")
	respHeader.WriteString(seg.TraceID)

	seg.Sampled = trace["Sampled"] != "0"
	if trace["Sampled"] == "?" {
		respHeader.WriteString(";Sampled=")
		respHeader.WriteString(strconv.Itoa(btoi(seg.Sampled)))
	}
	return respHeader.String()
}

func parseHeaders(h http.Header) map[string]string {
	m := map[string]string{}
	s := h.Get(headerTraceID)
	for _, c := range strings.Split(s, ";") {
		p := strings.SplitN(c, "=", 2)
		k := strings.TrimSpace(p[0])
		v := ""
		if len(p) > 1 {
			v = strings.TrimSpace(p[1])
		}
		m[k] = v
	}
	return m
}

func btoi(b bool) int {
	if b {
		return 1
	}
	return 0
}
