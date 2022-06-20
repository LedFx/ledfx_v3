package event

import (
	"testing"
)

func TestLog(t *testing.T) {
	handleLogEvent := func(e *LogEvent) {
		t.Log(e.Level, e.Msg)
	}
	id := SubscribeLog(handleLogEvent)
	InvokeLog("testLevel", "helloooo")
	UnsubscribeLog(id)
	InvokeLog("testAgain", "all is quiet")
}
