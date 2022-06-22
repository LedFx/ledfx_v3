package logger

import (
	"fmt"
	"ledfx/event"
	"os"
	"path/filepath"
	"runtime"
	"time"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
)

var Logger *logrus.Logger

func init() {
	Logger = logrus.New()
	Logger.SetOutput(os.Stderr)
	Logger.SetFormatter(&nested.Formatter{
		TimestampFormat: time.StampMilli,
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

// Hook log messages to fire internal ledfx events
type LogEventHook struct{}

func (l *LogEventHook) Levels() []logrus.Level {
	// everything up to info level (not debug or trace) should emit a logging event
	// MUST NOT include debug messages or events will go crazy in an infinite loop
	return []logrus.Level{0, 1, 2, 3, 4}
}

func (l *LogEventHook) Fire(e *logrus.Entry) error {
	event.Invoke(
		event.Log,
		map[string]interface{}{
			"level": e.Level.String(),
			"msg":   e.Message,
		},
	)
	return nil
}

func init() {
	Logger.AddHook(&LogEventHook{})
}
