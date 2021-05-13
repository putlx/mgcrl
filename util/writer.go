package util

import (
	"bufio"
	"io"
)

type Writer struct {
	w *bufio.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{bufio.NewWriter(w)}
}

func (w *Writer) Write(p []byte) (nn int, err error) {
	defer w.w.Flush()
	return w.w.Write(p)
}
