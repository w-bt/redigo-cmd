package logger

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

func Debugf(format string, args ...interface{}) {
	fields := getSourceInfoFields(3)
	fields["root"] = true
	WithFields(fields).Debugf(format, args...)
}

func Infof(format string, args ...interface{}) {
	fields := getSourceInfoFields(3)
	fields["root"] = true
	WithFields(fields).Infof(format, args...)
}

func Warnf(format string, args ...interface{}) {
	fields := getSourceInfoFields(3)
	fields["root"] = true
	WithFields(fields).Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	fields := getSourceInfoFields(3)
	fields["root"] = true
	WithFields(fields).Errorf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
	fields := getSourceInfoFields(3)
	fields["root"] = true
	WithFields(fields).Fatalf(format, args...)
}

func Error(args ...interface{}) {
	fields := getSourceInfoFields(3)
	fields["root"] = true
	WithFields(fields).Error(args...)
}

func IsDebugMode() bool {
	return loggerInstance.Level == logrus.DebugLevel
}

// WithFields provides an instance of loggerInstance annotated with the
// given fields.
func WithFields(fields Fields) Logger {
	srcInfo := getSourceInfoFields(3)
	if fields != nil {
		for key, val := range fields {
			srcInfo[key] = val
		}
	}
	withFields := loggerInstance.WithFields(srcInfo)
	if initFields != nil {
		withFields = withFields.WithFields(initFields)
	}
	return withFields
}

// Fields represents key-value pairs and can be used to
// provide additional context in logs
type Fields map[string]interface{}

// Logger represents a generic logging component
type Logger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Error(args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
}

func getSourceInfoFields(subtractStackLevels int) map[string]interface{} {
	file, line := getFileInfo(subtractStackLevels)
	m := map[string]interface{}{
		"f": fmt.Sprintf("%s:%d", file, line),
	}
	return m
}

func getFileInfo(subtractStackLevels int) (string, int) {
	_, file, line, _ := runtime.Caller(subtractStackLevels)
	return chopPath(file), line
}

// return the source filename after the last slash
func chopPath(original string) string {
	i := strings.LastIndex(original, "/")
	if i != -1 {
		return original[i+1:]
	}
	return original
}
