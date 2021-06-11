package util

import (
	"os"
)

type Writer struct {
	f string
}

func NewWriter(f string) *Writer {
	return &Writer{f}
}

func (w *Writer) Write(p []byte) (int, error) {
	f, err := os.OpenFile(w.f, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	return f.Write(p)
}
