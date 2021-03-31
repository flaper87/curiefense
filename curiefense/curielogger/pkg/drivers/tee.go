package drivers

import (
	"go.uber.org/atomic"
	"io"
)

type Tee struct {
	fanOuts []io.WriteCloser
	closed  *atomic.Bool
}

func NewTee(drivers []io.WriteCloser) io.WriteCloser {
	return &Tee{fanOuts: drivers, closed: atomic.NewBool(false)}
}

func (b *Tee) Write(p []byte) (n int, err error) {
	if b.closed.Load() {
		return 0, nil
	}
	for _, d := range b.fanOuts {
		if _, e := d.Write(p); e != nil {
			err = e
		}
	}
	return len(p), err
}

func (b *Tee) Close() error {
	b.closed.Store(true)
	var err error
	for _, d := range b.fanOuts {
		if e := d.Close(); e != nil {
			err = e
		}
	}
	return err
}
