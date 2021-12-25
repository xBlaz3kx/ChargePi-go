package logging

import (
	"fmt"
	graylog "github.com/gemnasium/logrus-graylog-hook/v3"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	log "github.com/sirupsen/logrus"
	lSyslog "github.com/sirupsen/logrus/hooks/syslog"
	"github.com/xBlaz3kx/ChargePi-go/data/settings"
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
)

func SetupLogs(loggingConfig settings.Logging, isDebug bool) {
	// Default (production) logging settings
	log.SetLevel(log.WarnLevel)
	log.SetFormatter(&log.JSONFormatter{})

	if isDebug {
		log.SetLevel(log.DebugLevel)
		log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	}

	logFormat := LogFormat(loggingConfig.Format)

	for _, logType := range loggingConfig.Type {

		switch LogType(logType) {
		case FileLogging:
			fileLogging(isDebug, "/var/logs/chargepi.log")
			break
		case RemoteLogging:
			remoteLogging(loggingConfig.Host, loggingConfig.Port, logFormat)
			break
		case ConsoleLogging:
			break
		}

	}
}

func remoteLogging(host string, port int, format LogFormat) {
	var (
		address = fmt.Sprintf("%s:%d", host, port)
	)

	switch format {
	case Gelf:
		hook := graylog.NewAsyncGraylogHook(address, map[string]interface{}{})
		defer hook.Flush()
		log.AddHook(hook)
		break
	case Json:
		break
	case Syslog:
		hook, err := lSyslog.NewSyslogHook(
			"tcp",
			address,
			syslog.LOG_WARNING,
			"chargePi",
		)
		if err == nil {
			log.AddHook(hook)
		}
		break
	}
}

func fileLogging(isDebug bool, path string) {
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

	log.AddHook(hook)
}
