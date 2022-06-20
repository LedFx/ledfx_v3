package event

import (
	"ledfx/logger"
	"ledfx/util"
	"time"

	"github.com/sirupsen/logrus"
)

var callbacksLog map[string]func(*LogEvent) = map[string]func(*LogEvent){}

type LogEvent struct {
	Type      string
	Timestamp time.Time
	Level     string
	Msg       string
}

func InvokeLog(level, msg string) {
	e := LogEvent{
		Type:      "Log",
		Timestamp: time.Now(),
		Level:     level,
		Msg:       msg,
	}
	for _, cb := range callbacksLog {
		cb(&e)
	}
}

func SubscribeLog(cb func(*LogEvent)) string {
	id := util.RandID()
	callbacksLog[id] = cb
	return id
}

func UnsubscribeLog(id string) {
	delete(callbacksLog, id)
}

// adding this event type to logger

type LogEventHook struct{}

func (l *LogEventHook) Levels() []logrus.Level {
	return []logrus.Level{0, 1, 2, 3, 4, 5, 6}
}

func (l *LogEventHook) Fire(e *logrus.Entry) error {
	InvokeLog(
		e.Level.String(),
		e.Message,
	)
	return nil
}

func init() {
	logger.Logger.AddHook(&LogEventHook{})
}
