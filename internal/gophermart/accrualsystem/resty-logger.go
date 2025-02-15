package accrualsystem

import (
	"go.uber.org/zap"
)

type RestyLogger struct {
	zapLogger *zap.SugaredLogger
}

func NewRestyLogger(logger *zap.Logger) *RestyLogger {
	return &RestyLogger{
		zapLogger: logger.Sugar(),
	}
}

func (l *RestyLogger) Errorf(format string, v ...interface{}) {
	l.zapLogger.Errorf(format, v...)
}
func (l *RestyLogger) Warnf(format string, v ...interface{}) {
	l.zapLogger.Warnf(format, v...)
}
func (l *RestyLogger) Debugf(format string, v ...interface{}) {
	l.zapLogger.Debugf(format, v...)
}
