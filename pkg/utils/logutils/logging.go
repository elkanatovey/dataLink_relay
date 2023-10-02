package logutils

import (
	"github.com/sirupsen/logrus"
)

func SetLogStyle() {
	logrus.SetReportCaller(true)
	jsonFormatter := &logrus.JSONFormatter{
		PrettyPrint:     true,
		TimestampFormat: "2006-01-02 15:04:05",
	}
	logrus.SetFormatter(jsonFormatter)
}
