package event

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"
)

/* inflexible and hardcoded, but fast and reliable... */

type EventType int

const (
	Log EventType = iota
	EffectRender
	EffectUpdate
)

func (et EventType) String() string {
	switch et {
	case Log:
		return "Log"
	case EffectRender:
		return "Effect Render"
	case EffectUpdate:
		return "Effect Update"
	default:
		return "Unknown"
	}
}

type Event struct {
	Timestamp time.Time
	Type      EventType
	Data      map[string]interface{}
}

type callbacks map[string]func(*Event)

var listeners map[EventType]callbacks = make(map[EventType]callbacks)
var mu sync.Mutex = sync.Mutex{}
var ErrInvalidEvent error = errors.New("invalid event data")

func Subscribe(et EventType, cb func(*Event)) (unsub func()) {
	// create callbacks if it doesn't exist
	if _, exists := listeners[et]; !exists {
		listeners[et] = make(callbacks)
	}

	// generate and store the callback with a unique id
	id := randID()
	listeners[et][id] = cb

	// return a func to unsubscribe
	unsub = func() {
		mu.Lock()
		defer mu.Unlock()
		delete(listeners[et], id)
		if len(listeners[et]) == 0 {
			delete(listeners, et)
		}
	}

	return unsub
}

func Invoke(et EventType, data map[string]interface{}) {
	// make sure event has the right keys in data
	var err error
	switch et {
	case Log:
		err = checkKeys(data, []string{"level", "msg"})
	case EffectRender:
		err = checkKeys(data, []string{"pixels"})
	case EffectUpdate:
		err = checkKeys(data, []string{"config", "id"})
	}

	// Do not invoke the event if it's missing keys
	if err != nil {
		return
	}

	// make event
	event := Event{
		Timestamp: time.Now(),
		Type:      et,
		Data:      data,
	}

	// invoke callbacks
	for _, cb := range listeners[et] {
		cb(&event)
	}
}

func checkKeys(m map[string]interface{}, keys []string) error {
	for _, key := range keys {
		if _, exists := m[key]; !exists {
			return ErrInvalidEvent
		}
	}
	return nil
}

func randID() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.Dump(b)
}
