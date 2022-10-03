package logger

import (
	"github.com/sirupsen/logrus"
)

type Logger interface {
	Debug(msg ...interface{})
	Debugf(format string, args ...interface{})
	Info(msg ...interface{})
	Infof(format string, args ...interface{})
	Warn(msg ...interface{})
	Warnf(format string, args ...interface{})
	Error(msg ...interface{})
	Errorf(format string, args ...interface{})
}

func Init() Logger {
	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})

	return log
}
