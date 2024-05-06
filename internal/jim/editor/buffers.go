package editor

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

type Buffer interface {
	io.Closer
	Load() error
	Save() (written int, err error)
	LinesInRange(lineRange LineRange) []*Line
	InsertText(position Point, text string) error
	DeleteText(position Point, length int) error
}

type FileBuffer struct {
	lines []*Line
	path  string
	file  *os.File
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
	scanner := bufio.NewScanner(f)
	lineNumber := int64(1)
	for scanner.Scan() {
		fb.lines = append(fb.lines, &Line{content: scanner.Text(), number: lineNumber})
		lineNumber += 1
	}
	if err := scanner.Err(); err != nil {
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

func (fb *FileBuffer) Save() (written int, err error) {
	if err := fb.ensureFile(); err != nil {
		return 0, err
	}

	totalWritten := 0
	for _, line := range fb.lines {
		written, err := fb.file.WriteString(line.content)
		totalWritten += written
		if err != nil {
			return totalWritten, err
		}
	}
	return totalWritten, nil
}

func (fb *FileBuffer) Close() error {
	if fb.file == nil {
		return nil
	}
	return fb.file.Close()
}

func (fb *FileBuffer) LinesInRange(lr LineRange) []*Line {
	return fb.lines[lr.start-1 : lr.end-1]
}

func (fb *FileBuffer) InsertText(p Point, text string) error {
	line := fb.lines[p.RowIndex()]
	line.content = line.content[:p.ColumnIndex()] + text + line.content[p.ColumnIndex():]
	return nil
}

func (fb *FileBuffer) DeleteText(p Point, length int) error {
	if length == 0 {
		return nil
	}

	line := fb.lines[p.RowIndex()]

	if length < 0 {
		line.content = line.content[:length+p.ColumnIndex()] + line.content[p.ColumnIndex():]
		return nil
	}

	line.content = line.content[:p.ColumnIndex()] + line.content[length+p.ColumnIndex():]
	return nil
}

func NewFileBuffer(path string) *FileBuffer {
	return &FileBuffer{path: path}
}
