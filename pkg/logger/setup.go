package logger

import (
	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/evalphobia/logrus_sentry"
	"github.com/getsentry/raven-go"
	"github.com/orandin/lumberjackrus"
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"log"
	"os"
	"runtime"
)

const RFC3339Milli = "2006-01-02T15:04:05.000Z"

var loggerInstance *logrus.Logger
var newRelicLogger *logrus.Logger
var accessLogFile lumberjack.Logger
var initFields logrus.Fields

func Setup(logLevel string, sentryEnabled bool, logPath string) {
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		level = logrus.WarnLevel
	}

	loggerInstance = getNewLoggerInstance(level, logPath, "application.log")
	newRelicLogger = getNewLoggerInstance(logrus.InfoLevel, logPath, "new_relic.log")
	log.SetOutput(loggerInstance.WriterLevel(logrus.DebugLevel))

	// Add error level logs to sentry
	if sentryEnabled {
		sentryHook, err := logrus_sentry.NewWithClientSentryHook(raven.DefaultClient, []logrus.Level{
			logrus.PanicLevel,
			logrus.FatalLevel,
			logrus.ErrorLevel,
		})

		if err == nil && sentryHook != nil {
			sentryHook.Timeout = 0
			sentryHook.StacktraceConfiguration.Enable = true
			sentryHook.StacktraceConfiguration.Skip = 7
			loggerInstance.Hooks.Add(sentryHook)
		}
	}

	// setup access log file
	accessLogFile = lumberjack.Logger{
		Filename:   logPath + "access.log",
		MaxSize:    200, // megabytes
		MaxBackups: 20,
		Compress:   true,  // disabled by default
		LocalTime:  false, // optional
		MaxAge:     7,
	}
}

func getNewLoggerInstance(level logrus.Level, logPath string, filename string) *logrus.Logger {
	instance := &logrus.Logger{
		Out:   os.Stdout,
		Hooks: make(logrus.LevelHooks),
		Level: level,
		Formatter: &nested.Formatter{
			TimestampFormat: RFC3339Milli,
			CustomCallerFormatter: func(frame *runtime.Frame) string {
				return ""
			},
		},
		ReportCaller: true,
	}

	lumberjackHook, err := lumberjackrus.NewHook(
		&lumberjackrus.LogFile{
			Filename:   logPath + filename,
			MaxSize:    200,
			MaxBackups: 20,
			Compress:   true,
			LocalTime:  false,
			MaxAge:     7,
		},
		level,
		&nested.Formatter{
			NoColors:        true,
			TimestampFormat: RFC3339Milli,
			CustomCallerFormatter: func(frame *runtime.Frame) string {
				return ""
			},
		},
		&lumberjackrus.LogFileOpts{},
	)
	instance.Hooks.Add(&LineNumberHook{})
	if err == nil {
		instance.Hooks.Add(lumberjackHook)
	}
	return instance
}

func init() {
	Setup("debug", false, "log/")
}

func GetAccessLogFile() io.WriteCloser {
	return &accessLogFile
}

func GetLoggerInstance() *logrus.Logger {
	return loggerInstance
}

func GetNewRelicLoggerInstance() *logrus.Logger {
	return newRelicLogger
}
