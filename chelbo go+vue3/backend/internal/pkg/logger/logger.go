package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func Init(environment string) {
	log = logrus.New()
	log.SetOutput(os.Stdout)

	if environment == "production" {
		log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		})
		log.SetLevel(logrus.InfoLevel)
	} else {
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
			ForceColors:     true,
		})
		log.SetLevel(logrus.DebugLevel)
	}
}

func Get() *logrus.Logger {
	if log == nil {
		Init("development")
	}
	return log
}

func Info(args ...interface{}) {
	Get().Info(args...)
}

func Infof(format string, args ...interface{}) {
	Get().Infof(format, args...)
}

func Error(args ...interface{}) {
	Get().Error(args...)
}

func Errorf(format string, args ...interface{}) {
	Get().Errorf(format, args...)
}

func Debug(args ...interface{}) {
	Get().Debug(args...)
}

func Debugf(format string, args ...interface{}) {
	Get().Debugf(format, args...)
}

func Warn(args ...interface{}) {
	Get().Warn(args...)
}

func Warnf(format string, args ...interface{}) {
	Get().Warnf(format, args...)
}

// Fatal logs with fatal level and exits
func Fatal(args ...interface{}) {
	Get().Fatal(args...)
}

// Fatalf logs with fatal level and exits
func Fatalf(format string, args ...interface{}) {
	Get().Fatalf(format, args...)
}
