package resp

import "io"

type Writer struct {
	Writer io.Writer
}

func (w *Writer) Write(v Value) error {
	_, err := w.Writer.Write(v.Marshal())
	return err
}
