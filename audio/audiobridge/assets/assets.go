package assets

import _ "embed"

//go:embed blankAlbumArt.png
var albumArt []byte

func BlankAlbumArt() []byte {
	return albumArt
}
