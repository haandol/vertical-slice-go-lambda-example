package instrument

import (
	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/haandol/vertical-slice-go-lambda-example/api/pkg/util/slogger"
	"github.com/pkg/errors"
)

func RecordGetClickStreamError(logger *slogger.Logger, span *xray.Segment, err error) {
	logger.Error(errors.Wrap(err, "fail to get click-stream").Error())
	if span != nil {
		_ = span.AddError(errors.WithStack(err))
	}
	increaseGetClickStreamErrorCount(logger)
}

func increaseGetClickStreamErrorCount(logger *slogger.Logger) {
	logger.Info("GetClickStreamError", "Error", 1)
}
