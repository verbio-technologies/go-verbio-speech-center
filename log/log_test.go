package log

import (
	"bytes"
	"regexp"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestNewLogger(t *testing.T) {
	logger := NewLogger("INFO")
	if logger == nil {
		t.Error("NewLogger returned nil")
	} else if logger.Level != logrus.InfoLevel {
		t.Errorf("Default level should be INFO, got %v", logger.Level)
	}
}

func TestSetLevel(t *testing.T) {
	logger := NewLogger("INFO")
	logger.SetLevel(logrus.DebugLevel)
	if logger.GetLevel() != logrus.DebugLevel {
		t.Errorf("Expected level DEBUG, got %v", logger.GetLevel())
	}
}

func TestLogging(t *testing.T) {
	tests := []struct {
		name     string
		level    logrus.Level
		logFunc  func(*logrus.Logger, ...interface{})
		message  string
		expected string
	}{
		{"Debug", logrus.DebugLevel, (*logrus.Logger).Debug, "debug message", `debug.*debug message`},
		{"Info", logrus.InfoLevel, (*logrus.Logger).Info, "info message", `info.*info message`},
		{"Warning", logrus.WarnLevel, (*logrus.Logger).Warning, "warning message", `warning.*warning message`},
		{"Error", logrus.ErrorLevel, (*logrus.Logger).Error, "error message", `error.*error message`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewLogger("INFO")
			logger.SetOutput(&buf)
			logger.SetLevel(tt.level)

			tt.logFunc(logger, tt.message)
			got := buf.String()
			if !regexp.MustCompile(tt.expected).MatchString(got) {
				t.Errorf("Expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestLogLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger("INFO")
	logger.SetOutput(&buf)
	logger.SetLevel(logrus.WarnLevel)

	logger.Debug("debug message")
	logger.Info("info message")
	if buf.String() != "" {
		t.Error("Expected no output for Debug/Info when level is WARNING")
	}

	logger.Warning("warning message")
	if buf.String() == "" {
		t.Errorf("Expected warning message, got nothing")
	}
}

func TestFormatf(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger("INFO")
	logger.SetOutput(&buf)

	logger.Infof("Hello %s", "world")
	expected := "info.*Hello world"
	got := buf.String()
	if !regexp.MustCompile(expected).MatchString(got) {
		t.Errorf("Expected %q, got %q", expected, got)
	}
}

func TestGetLevel(t *testing.T) {
	logger := NewLogger("INFO")
	if logger.GetLevel() != logrus.InfoLevel {
		t.Errorf("Default level should be INFO, got %v", logger.GetLevel())
	}
}
