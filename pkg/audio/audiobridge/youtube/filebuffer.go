package youtube

import (
	"bytes"
	"io/ioutil"
	"os"
)

type FileBuffer struct {
	buf    *bytes.Buffer
	offset int64
	closed bool
}

func NewFileBuffer(fi *os.File) (*FileBuffer, error) {
	defer fi.Close()
	fileBytes, err := ioutil.ReadAll(fi)
	if err != nil {
		return nil, err
	}
	return &FileBuffer{
		buf:    bytes.NewBuffer(fileBytes),
		offset: 0,
	}, nil
}

func (fb *FileBuffer) CurrentOffset() int64 {
	return fb.offset
}
func (fb *FileBuffer) Read(p []byte) (n int, err error) {
	if fb.closed {
		return 0, os.ErrClosed
	} else {
		n, err = fb.buf.Read(p)
		fb.offset += int64(n)
		return n, err
	}
}

func (fb *FileBuffer) Close() error {
	fb.closed = true
	fb.buf.Reset()
	fb.buf = &bytes.Buffer{}
	return nil
}
