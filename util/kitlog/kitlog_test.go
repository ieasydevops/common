package kitlog

import (
	"os"
	stdlog "log"
	kitlog "github.com/go-kit/kit/log"
	"testing"
)

func Test_kitlog(t *testing.T) {
	logger := kitlog.NewJSONLogger(kitlog.NewSyncWriter(os.Stdout))
	stdlog.SetOutput(kitlog.NewStdlibAdapter(logger))
	stdlog.Print("I sure like pie")
}