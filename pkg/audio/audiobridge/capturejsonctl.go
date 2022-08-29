package audiobridge

import (
	"encoding/json"
	"fmt"
)

type CaptureAction int

const (
	CaptureActionStop CaptureAction = iota
)

type CaptureCTLJSON struct {
	Action CaptureAction `json:"action"`
}

func (capctl CaptureCTLJSON) AsJSON() ([]byte, error) {
	return json.Marshal(&capctl)
}

func (j *JsonCTL) Capture(jsonData []byte) (err error) {
	conf := CaptureCTLJSON{}
	if err := json.Unmarshal(jsonData, &conf); err != nil {
		return fmt.Errorf("error unmarshalling JSON: %w", err)
	}

	switch conf.Action {
	case CaptureActionStop:
		return j.w.br.Controller().Local().QuitCapture()
	}

	return fmt.Errorf("unknown action '%d'", conf.Action)
}
