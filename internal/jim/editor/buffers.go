package editor

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
)

const (
	MaxLineLength = 10 * 1024 * 1024
)

type Buffer interface {
	fmt.Stringer
	io.Closer
	io.ReaderFrom
	io.Writer
	Load() error
	Save() (written int, err error)
	Clear()
	LinesInRange(lineRange LineRange) []*Line
	InsertText(position Point, text string) error
	DeleteText(position Point, length int) error
}

type MemoryBuffer struct {
	logger *slog.Logger
	lines  []*Line
}

func (mb *MemoryBuffer) LinesInRange(lr LineRange) []*Line {
	return mb.lines[lr.start-1 : lr.end]
}

func (mb *MemoryBuffer) InsertText(p Point, text string) error {
	line := mb.lines[p.RowIndex()]
	line.content = line.content[:p.ColumnIndex()] + text + line.content[p.ColumnIndex():]
	return nil
}

func (mb *MemoryBuffer) DeleteText(p Point, length int) error {
	if length == 0 {
		return nil
	}

	line := mb.lines[p.RowIndex()]

	if length < 0 {
		line.content = line.content[:length+p.ColumnIndex()] + line.content[p.ColumnIndex():]
		return nil
	}

	line.content = line.content[:p.ColumnIndex()] + line.content[length+p.ColumnIndex():]
	return nil
}

func (mb *MemoryBuffer) Clear() {
	mb.lines = []*Line{{}}
}

func (mb *MemoryBuffer) String() string {
	var sb strings.Builder
	for _, line := range mb.lines {
		sb.WriteString(line.content + "\n")
	}
	return sb.String()
}

func (mb *MemoryBuffer) Write(p []byte) (n int, err error) {
	panic("not implemented. use ReadFrom")
}

func (mb *MemoryBuffer) ReadFrom(r io.Reader) (n int64, err error) {
	mb.logger.Info("ReadFrom:", "reader", r)
	scanner := bufio.NewScanner(r)
	// TODO: refactor to use a bufio reader
	buf := make([]byte, 0, MaxLineLength)
	scanner.Buffer(buf, MaxLineLength)
	var totalRead int64
	lineNumber := int64(1)
	for scanner.Scan() {
		content := scanner.Text()
		mb.logger.Debug("appending line: ", "content", content, "number", lineNumber)
		mb.lines = append(mb.lines, &Line{content: content, number: lineNumber})
		lineNumber += 1
		totalRead += int64(len(scanner.Bytes()))
	}
	if err := scanner.Err(); err != nil {
		return totalRead, fmt.Errorf("memory buffer read: %w", err)
	}
	return
}

func (MemoryBuffer) Close() error {
	// Noop
	return nil
}

func (MemoryBuffer) Load() error {
	// Noop
	return nil
}

func (MemoryBuffer) Save() (written int, err error) {
	// Noop
	return 0, nil
}

type FileBuffer struct {
	mbuf   *MemoryBuffer
	logger *slog.Logger
	path   string
	file   *os.File
}

func (fb *FileBuffer) String() string {
	return fb.mbuf.String()
}

func (fb *FileBuffer) ReadFrom(r io.Reader) (n int64, err error) {
	return fb.mbuf.ReadFrom(r)
}

func (fb *FileBuffer) Write(p []byte) (n int, err error) {
	return fb.mbuf.Write(p)
}

func (fb *FileBuffer) LinesInRange(lineRange LineRange) []*Line {
	return fb.mbuf.LinesInRange(lineRange)
}

func (fb *FileBuffer) InsertText(position Point, text string) error {
	return fb.mbuf.InsertText(position, text)
}

func (fb *FileBuffer) DeleteText(position Point, length int) error {
	return fb.mbuf.DeleteText(position, length)
}

func (fb *FileBuffer) Clear() {
	fb.mbuf.Clear()
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
	fb.mbuf = NewMemoryBuffer(fb.logger)

	if _, err := io.Copy(fb.mbuf, fb.file); err != nil {
		return err
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
	f := fb.file
	if err := fb.ensureFile(); err != nil {
		return 0, err
	}

	// TODO: make this safer (backups? atomic writes?)
	if err := f.Truncate(0); err != nil {
		return 0, fmt.Errorf("save truncate: %w", err)
	}
	if _, err := f.Seek(0, 0); err != nil {
		return 0, fmt.Errorf("save seek: %w", err)
	}

	// TODO: write incrementally
	return f.WriteString(fb.mbuf.String())
}

func (fb *FileBuffer) Close() error {
	if fb.file == nil {
		return nil
	}
	return fb.file.Close()
}

func NewFileBuffer(path string, logger *slog.Logger) *FileBuffer {
	return &FileBuffer{
		logger: logger,
		mbuf:   NewMemoryBuffer(logger),
		path:   path,
	}
}

func NewMemoryBuffer(logger *slog.Logger) *MemoryBuffer {
	return &MemoryBuffer{
		logger: logger,
		lines:  []*Line{},
	}
}
