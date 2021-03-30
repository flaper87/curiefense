package drivers

import (
	"bufio"
	"io"
	"os"
)

type BufferedStdout struct {
	buff *bufio.Writer
}

func NewBufferedStdout() io.WriteCloser {
	return &BufferedStdout{buff: bufio.NewWriterSize(os.Stdout, 1<<20)}
}

func (b *BufferedStdout) Write(p []byte) (n int, err error) {
	return b.buff.Write(p)
}

func (b *BufferedStdout) Close() error {
	b.buff.Flush()
	return nil
}
