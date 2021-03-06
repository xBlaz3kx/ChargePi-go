package logging

import (
	"fmt"
	graylog "github.com/gemnasium/logrus-graylog-hook/v3"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	log "github.com/sirupsen/logrus"
	lSyslog "github.com/sirupsen/logrus/hooks/syslog"
	"github.com/xBlaz3kx/ChargePi-go/internal/models/settings"
	"log/syslog"
	"time"
)

type (
	LogType   string
	LogFormat string
)

const (
	RemoteLogging  = LogType("remote")
	FileLogging    = LogType("file")
	ConsoleLogging = LogType("console")

	Syslog = LogFormat("syslog")
	Gelf   = LogFormat("gelf")
	Json   = LogFormat("json")

	logFilePath = "/var/log/chargepi/chargepi.log"
)

// Setup set up all logs
func Setup(logger *log.Logger, loggingConfig settings.Logging, isDebug bool) {
	var (
		// Default (production) logging settings
		logLevel                = log.WarnLevel
		formatter log.Formatter = &log.JSONFormatter{}
		logFormat               = LogFormat(loggingConfig.Format)
	)

	if isDebug {
		logLevel = log.DebugLevel
	}

	logger.SetFormatter(formatter)
	logger.SetLevel(logLevel)

	for _, logType := range loggingConfig.Type {
		switch LogType(logType) {
		case FileLogging:
			fileLogging(logger, isDebug, logFilePath)
			break
		case RemoteLogging:
			remoteLogging(logger, loggingConfig.Host, loggingConfig.Port, logFormat)
			break
		case ConsoleLogging:
			break
		}
	}
}

// remoteLogging sends logs remotely to Graylog or any Syslog receiver.
func remoteLogging(logger *log.Logger, host string, port int, format LogFormat) {
	var (
		address = fmt.Sprintf("%s:%d", host, port)
		hook    log.Hook
		err     error
	)

	switch format {
	case Gelf:
		graylogHook := graylog.NewAsyncGraylogHook(address, map[string]interface{}{})
		defer graylogHook.Flush()
		hook = graylogHook
		break
	case Json:
		break
	case Syslog:
		hook, err = lSyslog.NewSyslogHook(
			"tcp",
			address,
			syslog.LOG_WARNING,
			"chargePi",
		)
		break
	default:
		return
	}

	if err == nil {
		logger.AddHook(hook)
	}
}

// fileLogging sets up the logging to file.
func fileLogging(logger *log.Logger, isDebug bool, path string) {
	writer, err := rotatelogs.New(
		path+".%Y%m%d%H%M",
		rotatelogs.WithLinkName(path),
		rotatelogs.WithMaxAge(time.Duration(86400)*time.Second),
		rotatelogs.WithRotationTime(time.Duration(604800)*time.Second),
	)
	if err != nil {
		return
	}

	writerMap := make(lfshook.WriterMap)
	writerMap[log.InfoLevel] = writer
	writerMap[log.ErrorLevel] = writer

	if isDebug {
		writerMap[log.DebugLevel] = writer
	}

	hook := lfshook.NewHook(
		writerMap,
		&log.JSONFormatter{},
	)

	logger.AddHook(hook)
}
