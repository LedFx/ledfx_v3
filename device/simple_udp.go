package device

/*
A simple udp device for testing purposes.
This will be removed with the addition of virtuals and a proper device framework
*/

type SimpleUDP struct {
	Name       string
	PixelCount int
	IP         string
	Port       int
}
