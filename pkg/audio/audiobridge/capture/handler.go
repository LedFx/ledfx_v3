package capture

import (
	"fmt"

	"github.com/LedFx/ledfx/pkg/audio"
	log "github.com/LedFx/ledfx/pkg/logger"
	"github.com/gen2brain/malgo"
)

type Handler struct {
	*malgo.Device
	byteWriter *audio.AsyncMultiWriter
	stopped    bool
}

func NewHandler(id string, byteWriter *audio.AsyncMultiWriter) (h *Handler, err error) {
	deviceInfo, deviceType, err := audio.GetDeviceByID(id)
	if err != nil {
		return nil, err
	}
	log.Logger.WithField("context", "Local Capture Init").Debugf("Getting info for device '%s'...", deviceInfo.Name)

	config := malgo.DefaultDeviceConfig(deviceType)
	config.SampleRate = uint32(audio.SampleRate)
	config.PeriodSizeInFrames = uint32(audio.FramesPerBuffer)
	config.Capture.DeviceID = deviceInfo.ID.Pointer()
	config.Capture.Channels = 1
	config.Capture.Format = malgo.FormatS16

	h = &Handler{
		byteWriter: byteWriter,
	}

	callbacks := malgo.DeviceCallbacks{
		Data: func(pOutputSample []byte, pInputSamples []byte, framecount uint32) {
			h.byteWriter.Write(pInputSamples)
		},
		Stop: func() {},
	}

	log.Logger.WithField("context", "Local Capture Init").Debug("Initialising device...")
	h.Device, err = malgo.InitDevice(audio.Context.Context, config, callbacks)
	if err != nil {
		return nil, fmt.Errorf("error initialising stream: %w", err)
	}

	log.Logger.WithField("context", "Local Capture Init").Debugf("Starting stream...")
	if err = h.Device.Start(); err != nil {
		return nil, fmt.Errorf("error starting stream: %w", err)
	}

	return h, nil
}

func (h *Handler) Quit() {
	h.stopped = true
	log.Logger.WithField("context", "Capture Handler").Warnf("Stopping stream...")
	h.Device.Stop()
}

func (h *Handler) Stopped() bool {
	return h.stopped
}
