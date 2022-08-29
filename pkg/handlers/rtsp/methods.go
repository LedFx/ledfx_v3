//go:generate stringer -type=Method methods.go
package rtsp

type Method int

const (
	Describe Method = iota
	Announce
	Get_Parameter
	Options
	Play
	Pause
	Record
	Redirect
	Setup
	Set_Parameter
	Teardown
	Flush
)
