package audio

/*
import (
	"fmt"
	"ledfx/config"
	"ledfx/logger"
	"time"

	"github.com/gordonklaus/portaudio"
	"github.com/simonassank/aubio-go"
)

var frameCount int = 0
var pvoc *aubio.PhaseVoc
var melbank *aubio.FilterBank
var onset *aubio.Onset

func CaptureDemo() {
	// REMOVE THIS ONCE WE HAVE CONFIG VALIDATION
	config.GlobalConfig.Audio.FrameRate = 60

	if err := portaudio.Initialize(); err != nil {
		logger.Logger.Error(err)
		return
	}
	defer portaudio.Terminate()
	// match our config device to a real portaudio device
	di, err := GetPaDeviceInfo(config.GlobalConfig.Audio.Device)
	if err != nil {
		logger.Logger.Errorf("Audio device does not exist")
		return
	}
	// frames per buffer
	fpb := int(di.DefaultSampleRate) / config.GlobalConfig.Audio.FrameRate

	// phase vocoder
	pvoc, err = aubio.NewPhaseVoc(fftSize, uint(fpb))
	defer pvoc.Free()
	if err != nil {
		logger.Logger.Error(err)
		return
	}

	// filterbank
	melbank = aubio.NewFilterBank(40, fftSize)
	melbank.SetMelCoeffsSlaney(uint(di.DefaultSampleRate))

	// onset
	onset, err = aubio.NewOnset(aubio.HFC, uint(fftSize), uint(fpb), uint(di.DefaultSampleRate))
	if err != nil {
		logger.Logger.Error(err)
		return
	}

	p := portaudio.StreamParameters{
		Input: portaudio.StreamDeviceParameters{
			Device:   di,
			Channels: 1,
		},
		SampleRate:      di.DefaultSampleRate,
		FramesPerBuffer: fpb,
	}
	s, err := portaudio.OpenStream(p, audioSampleCallback)
	if err != nil {
		logger.Logger.Error(err)
		return
	}
	logger.Logger.Info("Starting stream, collecting audio...")
	logger.Logger.Info("Clap your hands!")
	err = s.Start()
	if err != nil {
		logger.Logger.Error(err)
		return
	}
	defer s.Close()
	time.Sleep(10 * time.Second)
	err = s.Stop()
	if err != nil {
		logger.Logger.Error(err)
		return
	}
	logger.Logger.Info("Stopped stream")
	// logger.Logger.Infof("Captured %d frames in 1 second, expected %d", frameCount, 60)
}

func audioSampleCallback(in Buffer) {
	frameCount += 1
	buf := aubio.NewSimpleBufferData(uint(len(in)), in.AsFloat64())
	defer buf.Free()
	pvoc.Do(buf)
	melbank.Do(pvoc.Grain())
	// fmt.Println(melbank.Buffer().Slice())
	onset.Do(buf)
	if onset.Buffer().Slice()[0] != 0 {
		fmt.Println("nice clap!")
	}
}
*/
