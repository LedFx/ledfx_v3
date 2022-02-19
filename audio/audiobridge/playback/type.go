package playback

type Handler interface {
	Identifier() string
	Quit()
	Write(p []byte) (n int, err error)
}
