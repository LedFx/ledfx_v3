package multicloser

import "io"

type multiCloser struct {
	io.Writer
	cs []io.Closer
}

func New(writers ...io.Writer) io.WriteCloser {
	m := &multiCloser{Writer: io.MultiWriter(writers...)}
	for _, w := range writers {
		if c, ok := w.(io.Closer); ok {
			m.cs = append(m.cs, c)
		}
	}
	return m
}

func (m *multiCloser) Close() error {
	var first error
	for _, c := range m.cs {
		if err := c.Close(); err != nil && first == nil {
			first = err
		}
	}
	return first
}
