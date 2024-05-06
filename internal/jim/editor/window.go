package editor

import (
	"fmt"
	"log/slog"
	"strings"
)

type Window struct {
	logger       *slog.Logger
	buffer       Buffer
	visibleLines LineRange
	cursor       Point
	width        int
	height       int
}

func NewWindow(width int, height int, logger *slog.Logger) *Window {
	return &Window{
		logger:       logger,
		buffer:       nil,
		cursor:       Point{1, 1},
		visibleLines: LineRange{1, int64(height)},
		width:        width,
		height:       height,
	}
}

func (w *Window) MoveCursorRelative(deltaRow int, deltaColumn int) {
	newRow := w.cursor.row + deltaRow
	newColumn := w.cursor.column + deltaColumn
	if newRow < 1 {
		newRow = 1
		w.shiftVisibleLines(int64(deltaRow))
	}
	if newRow > w.height {
		newRow = w.height
		w.shiftVisibleLines(int64(deltaRow))
	}
	if newColumn < 1 {
		newColumn = 1
	}
	// FIX:: This should actually be based on the current line length
	if newColumn > w.width {
		newColumn = w.width
	}

	w.MoveCursor(Point{newRow, newColumn})
}

func (w *Window) InsertText(p Point, text string) error {
	if err := w.buffer.InsertText(p, text); err != nil {
		return err
	}
	w.MoveCursorRelative(0, len(text))
	return nil
}

func (w *Window) DeleteText(p Point, length int) error {
	if err := w.buffer.DeleteText(p, length); err != nil {
		return err
	}
	if length < 0 {
		w.MoveCursorRelative(0, length)
	}
	return nil
}

func (w *Window) shiftVisibleLines(n int64) {
	w.logger.Debug(fmt.Sprintf("shifting visible lines by n: %d (%+v)", n, w.visibleLines))
	if w.visibleLines.start+n < 1 {
		n = -w.visibleLines.start + 1
		w.logger.Debug(fmt.Sprintf("went to zero so shifted to: %d", n))
		if n == 0 {
			return
		}
	}
	w.visibleLines = w.visibleLines.ShiftBy(n)
}

func (w *Window) MoveCursor(point Point) {
	w.cursor = point
}

func (w *Window) LoadBuffer(b Buffer) error {
	w.buffer = b
	return b.Load()
}

func (w *Window) CurrentPosition() Point {
	return Point{
		// TODO: figure out types here
		row: int(w.visibleLines.start + int64(w.cursor.row) - 1),
		// TODO: Handle visible column offset
		column: w.cursor.column,
	}
}

func (w *Window) Render() string {
	var sb strings.Builder
	for _, line := range w.buffer.LinesInRange(w.visibleLines) {
		sb.WriteString(line.content + "\r\n")
	}
	return sb.String()
}
