package logger

import (
	"github.com/sirupsen/logrus"
	"gitlab.com/sixbell/proyectos/brasil/claro/vas/elastic-query-export/flags"
	"io"
	"log"
	"os"
)

// Event stores messages to log later, from our standard interface
type Event struct {
	id      int
	message string
}

// StandardLogger enforces specific log message formats
type StandardLogger struct {
	*logrus.Logger
}

var confFlags *flags.Flags

// NewLogger initializes the standard logger
func NewLogger(conf *flags.Flags) *StandardLogger {
	var file *os.File
	var baseLogger = logrus.New()
	confFlags = conf

	var standardLogger = &StandardLogger{baseLogger}

	if conf.LogFile {
		standardLogger.SetOutput(file)
	}

	switch conf.LogLevel {
	case "debug":
		standardLogger.SetLevel(logrus.DebugLevel)
	case "info":
		standardLogger.SetLevel(logrus.InfoLevel)
	case "warn":
		standardLogger.SetLevel(logrus.WarnLevel)
	case "error":
		standardLogger.SetLevel(logrus.ErrorLevel)
	case "fatal":
		standardLogger.SetLevel(logrus.FatalLevel)
	case "panic":
		standardLogger.SetLevel(logrus.PanicLevel)
	default:
		standardLogger.SetLevel(logrus.InfoLevel)
	}

	if conf.LogFormat == "json" {
		standardLogger.Formatter = &logrus.JSONFormatter{}
	} else {
		standardLogger.Formatter = &logrus.TextFormatter{}
	}

	if conf.LogFile {
		closeFile(file)
	}
	return standardLogger
}

// Declare variables to store log messages as new Events
var (
	invalidArgMessage      = Event{1, "Invalid arg: %s"}
	invalidArgValueMessage = Event{2, "Invalid value for argument: %s: %v"}
	missingArgMessage      = Event{3, "Missing arg: %s"}
	infoMessage            = Event{4, " %s"}
	debugMessage           = Event{5, " %s"}
)

// InvalidArg is a standard error message
func (l *StandardLogger) InvalidArg(argumentName string) {
	logToFile(l)
	l.Errorf(invalidArgMessage.message, argumentName)
}

// InvalidArgValue is a standard error message
func (l *StandardLogger) InvalidArgValue(argumentName string, argumentValue string) {
	logToFile(l)
	l.Errorf(invalidArgValueMessage.message, argumentName, argumentValue)
}

// MissingArg is a standard error message
func (l *StandardLogger) MissingArg(argumentName string) {
	logToFile(l)
	l.Errorf(missingArgMessage.message, argumentName)
}

// Info is a standard info message
func (l *StandardLogger) Info(message string) {
	logToFile(l)
	l.Infof(infoMessage.message, message)
}

// Debug is a standard debug message
func (l *StandardLogger) Debug(message string) {
	logToFile(l)
	l.Debugf(debugMessage.message, message)
}

// Log is a standard log message
func (l *StandardLogger) Log(message string) {
	logToFile(l)
	l.Log(message)
}

func openFile() *os.File {
	file, err := os.OpenFile("elastic-query-export.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	return file
}

func closeFile(file *os.File) {
	err := file.Close()
	if err != nil {
		return
	}
}

func logToFile(l *StandardLogger) {
	if confFlags.LogFile {
		var file = openFile()
		mw := io.MultiWriter(os.Stdout, file)
		l.SetOutput(mw)
		closeFile(file)
	}
}
