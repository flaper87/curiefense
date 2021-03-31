package drivers

import (
	"io"
)

type Tee struct {
	fanOuts []io.WriteCloser
}

func NewTee(drivers []io.WriteCloser) io.WriteCloser {
	return &Tee{fanOuts: drivers}
}

func (b *Tee) Write(p []byte) (n int, err error) {
	for _, d := range b.fanOuts {
		if _, e := d.Write(p); e != nil {
			err = e
		}
	}
	return len(p), err
}

func (b *Tee) Close() error {
	return nil
}
