package logger

import (
	"fmt"
	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

var Logger *logrus.Logger

func init() {
	Logger = logrus.New()
	Logger.SetLevel(logrus.InfoLevel)
	Logger.SetNoLock()
	Logger.SetOutput(os.Stderr)
	Logger.SetFormatter(&nested.Formatter{
		FieldsOrder:     []string{"component", "category"},
		TimestampFormat: time.Stamp,
		HideKeys:        true,
		ShowFullLevel:   true,
		TrimMessages:    true,
		CallerFirst:     true,
		CustomCallerFormatter: func(frame *runtime.Frame) string {
			return fmt.Sprintf(" [%s:%d]", filepath.Base(frame.File), frame.Line)
		},
	})
	Logger.SetReportCaller(true)
}
