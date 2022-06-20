package event

import (
	"ledfx/color"
	"ledfx/util"
	"time"
)

var callbacksEffectRender map[string]func(*EffectRenderEvent) = map[string]func(*EffectRenderEvent){}

type EffectRenderEvent struct {
	Type      string
	Timestamp time.Time
	Pixels    color.Pixels
}

func InvokeEffectRender(p color.Pixels) {
	e := EffectRenderEvent{
		Type:      "Effect Render",
		Timestamp: time.Now(),
		Pixels:    p,
	}
	for _, cb := range callbacksEffectRender {
		cb(&e)
	}
}

func SubscribeEffectRender(cb func(*EffectRenderEvent)) string {
	id := util.RandID()
	callbacksEffectRender[id] = cb
	return id
}

func UnsubscribeEffectRender(id string) {
	delete(callbacksEffectRender, id)
}
