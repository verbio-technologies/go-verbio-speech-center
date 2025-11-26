package log

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

var (
	Logger = NewLogger("fatal")
)

type myFormatter struct {
	logrus.TextFormatter
	serviceName string
	id          string
}

func (m *myFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var message string
	if len(m.id) == 0 {
		message = fmt.Sprintf("[%s] %s", m.serviceName, entry.Message)
	} else {
		message = fmt.Sprintf("[%s] [%s] %s", m.serviceName, m.id, entry.Message)
	}

	entry.Message = message
	return m.TextFormatter.Format(entry)

}

func getTextFormatter() *logrus.TextFormatter {
	return &logrus.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	}
}

func createLogger(logLevel string, formatter logrus.Formatter) *logrus.Logger {
	logger := logrus.New()
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		logrus.Fatalf("Not a valid LogLevel [%s]", logLevel)
	}
	logger.SetLevel(level)
	logger.SetFormatter(formatter)
	logger.SetReportCaller(false)
	logger.Out = os.Stderr

	return logger
}

func NewLoggerWithNameAndId(name, id, logLevel string) *logrus.Logger {
	formatter := &myFormatter{
		TextFormatter: *getTextFormatter(),
		serviceName:   name,
		id:            id,
	}
	return createLogger(logLevel, formatter)
}

func NewLogger(logLevel string) *logrus.Logger {
	return createLogger(logLevel, getTextFormatter())
}

func InitLogger(logLevel string) {
	Logger = NewLogger(logLevel)
}

func InitTestLogger() {
	InitLogger("ERROR")
}
