package playback

type Handler interface {
	Write(p []byte) (n int, err error)

	Identifier() string
	Device() string

	SampleRate() int
	CurrentBufferSize() int
	NumChannels() int8

	Quit()
}
