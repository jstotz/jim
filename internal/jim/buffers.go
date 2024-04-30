package jim

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

type Buffer interface {
	io.ReadWriteCloser
	Load() error
}

type FileBuffer struct {
	bytes.Buffer
	path string
	file *os.File
}

func (fb *FileBuffer) Load() error {
	f, err := os.OpenFile(fb.path, os.O_RDWR, 0666)
	if os.IsNotExist(err) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("load file buffer: %w", err)
	}

	fb.file = f
	if _, err := io.Copy(fb, fb.file); err != nil {
		return fmt.Errorf("load file buffer: %w", err)
	}

	return nil
}

func (fb *FileBuffer) ensureFile() error {
	if fb.file != nil {
		return nil
	}
	f, err := os.Create(fb.path)
	if err != nil {
		return fmt.Errorf("file buffer: %w", err)
	}
	fb.file = f
	return nil
}

func (fb *FileBuffer) Save() (written int64, err error) {
	if err := fb.ensureFile(); err != nil {
		return 0, err
	}

	return io.Copy(fb.file, fb)
}

func (fb *FileBuffer) Close() error {
	if fb.file == nil {
		return nil
	}
	return fb.file.Close()
}

func NewFileBuffer(path string) *FileBuffer {
	return &FileBuffer{path: path}
}
