package mlog

import (
	"github.com/Sirupsen/logrus"
	"testing"
)

func Test_Logrus_New_Instance(t *testing.T) {
	origLogger.Formatter = &logrus.TextFormatter{
		DisableColors: false,
	}
	// The default logging level should be "info".
	Debug("This debug-level line should not show up in the output.")
	Infof("This %s-level line should show up in the output.", "info")

}

func Test_Logrus_Init(t *testing.T) {
	origLogger.Formatter = &logrus.TextFormatter{
		DisableColors: true,
	}
	origLogger.Level = logrus.DebugLevel
	// The default logging level should be "info".
	Debug("This debug-level line should not show up in the output.")
	Infof("This %s-level line should show up in the output.", "info")

	With("a", "100").Infof("aaa: %v", "asdfasdf")

}





