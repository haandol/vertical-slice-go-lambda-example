package instrument

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/haandol/vertical-slice-go-lambda-example/api/pkg/util/slogger"
	"github.com/pkg/errors"
)

func RecordBadInputError(logger *slogger.Logger, span *xray.Segment, err error) {
	logger.Error("bad input error", "err", err)
	logger.Error(errors.WithStack(err).Error())

	if span != nil {
		_ = span.AddError(errors.WithStack(err))
	}
}

func RecordError(logger *slogger.Logger, span *xray.Segment, err error) {
	logger.Error(errors.WithStack(err).Error())

	if span != nil {
		_ = span.AddError(errors.WithStack(err))
	}
}

func RecordRequest(logger *slogger.Logger, span *xray.Segment, request interface{}) {
	logger.Info(fmt.Sprintf("request: %v", request))
	value, err := json.Marshal(request)
	if err != nil {
		logger.Error("json marshal error", "err", err)
		return
	}

	if span != nil {
		_ = span.AddMetadata("request", string(value))
	}
}

func SetParamsToSpanAttr(logger *slogger.Logger, span *xray.Segment, params interface{}) {
	logger.Error("error on record data to database", "params", params)
	value, err := json.Marshal(params)
	if err != nil {
		logger.Error("json marshal error", "err", err)
		return
	}

	if span != nil {
		_ = span.AddMetadata("params", string(value))
	}
}
