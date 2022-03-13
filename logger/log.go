package logger

import (
	"os"

	log "github.com/sirupsen/logrus"
)

type AppLogger struct {
	*log.Logger
}

var Logger = NewLogger()

func NewLogger() *AppLogger {
	logger := log.New()

	logger.SetFormatter(&log.TextFormatter{})
	logger.SetOutput(os.Stdout)

	logger.SetLevel(log.InfoLevel)

	logger.Info("logger instance created")
	return &AppLogger{Logger: logger}
}

func (l *AppLogger) LogIfErr(err error) {
	if err != nil {
		l.Trace(err.Error())
	}
}
