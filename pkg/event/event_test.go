package event

import (
	"testing"
)

func TestEvents(t *testing.T) {
	handleLogEvent := func(e *Event) {
		t.Log(e.Data["level"], e.Data["msg"])
	}
	unsub := Subscribe(Log, handleLogEvent)
	Invoke(Log, map[string]interface{}{"level": "testing level", "msg": "testing message"})
	unsub()
	Invoke(Log, map[string]interface{}{"level": "testing level", "msg": "testing message again"})
}
