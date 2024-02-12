package instrument

import (
	"github.com/aws/aws-xray-sdk-go/xray"
	"github.com/haandol/vertical-slice-go-lambda-example/api/pkg/util/slogger"
	"github.com/pkg/errors"
)

func RecordCreateClickEventSuccess(logger *slogger.Logger) {
	logger.Info("create click-event success")
	increaseCreateClickEventSuccessCount(logger)
}

func RecordCreateClickEventError(logger *slogger.Logger, span *xray.Segment, err error) {
	logger.Error(errors.Wrap(err, "fail to create click-event").Error())
	if span != nil {
		_ = span.AddError(errors.WithStack(err))
	}
	increaseCreateClickEventErrorCount(logger)
}

func increaseCreateClickEventSuccessCount(logger *slogger.Logger) {
	logger.Info("CreateClickEvent", "Success", 1)
}

func increaseCreateClickEventErrorCount(logger *slogger.Logger) {
	logger.Info("CreateClickEvent", "Error", 1)
}
