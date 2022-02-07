package audiobridge

import (
	"encoding/json"
	"fmt"
)

type BridgeJSONWrapper struct {
	br *Bridge
}

func (br *Bridge) NewJSONWrapper() *BridgeJSONWrapper {
	return &BridgeJSONWrapper{
		br: br,
	}
}

func (w *BridgeJSONWrapper) StartAirPlayInput(jsonData []byte) (err error) {
	conf := AirPlayInputJSON{}
	if err := json.Unmarshal(jsonData, &conf); err != nil {
		return fmt.Errorf("error unmarshalling JSON: %w", err)
	}

	if err := w.br.StartAirPlayInput(conf.Name, conf.Port, conf.Verbose); err != nil {
		return fmt.Errorf("error starting AirPlay input (AirPlay Server): %w", err)
	}

	return nil
}

func (w *BridgeJSONWrapper) AddAirPlayOutput(jsonData []byte) (err error) {
	conf := AirPlayOutputJSON{}
	if err := json.Unmarshal(jsonData, &conf); err != nil {
		return fmt.Errorf("error unmarshalling JSON: %w", err)
	}
	if err := w.br.AddAirPlayOutput(conf.SearchKey, conf.SearchType, conf.Verbose); err != nil {
		return fmt.Errorf("error adding AirPlay output (AirPlay Client): %w", err)
	}
	return nil
}

func (w *BridgeJSONWrapper) StartLocalInput(jsonData []byte) (err error) {
	conf := LocalInputJSON{}
	if err := json.Unmarshal(jsonData, &conf); err != nil {
		return fmt.Errorf("error unmarshalling JSON: %w", err)
	}
	if err := w.br.StartLocalInput(*conf.AudioDevice, conf.Verbose); err != nil {
		return fmt.Errorf("error starting local input (capture): %w", err)
	}
	return nil
}

func (w *BridgeJSONWrapper) AddLocalOutput(jsonData []byte) (err error) {
	conf := LocalOutputJSON{}
	if err := json.Unmarshal(jsonData, &conf); err != nil {
		return fmt.Errorf("error unmarshalling JSON: %w", err)
	}
	if err := w.br.AddLocalOutput(conf.Verbose); err != nil {
		return fmt.Errorf("error starting local output (playback): %w", err)
	}
	return nil
}
