// Package logger
package logger

import (
	"github.com/sirupsen/logrus"
	"time"
)

func SetJSONFormatter() {
	logrus.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
	})
}

func Setup() {
	SetJSONFormatter()
	logrus.SetLevel(logrus.InfoLevel)

}
