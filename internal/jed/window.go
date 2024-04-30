package jed

import (
	"io"
	"strings"
)

type Window struct {
	buffer Buffer
	cursor Point
	width  int
	height int
}

func NewWindow(width int, height int) *Window {
	return &Window{
		buffer: nil,
		cursor: Point{1, 1},
		width:  width,
		height: height,
	}
}

func (w *Window) MoveCursorRelative(deltaRow int, deltaColumn int) {
	newRow := w.cursor.row + deltaRow
	newColumn := w.cursor.column + deltaColumn
	if newRow < 1 {
		newRow = 1
	}
	if newRow > w.height {
		newRow = w.height
	}
	if newColumn < 1 {
		newColumn = 1
	}
	// FIX:: This should actually be based on the current line length
	if newColumn > w.width {
		newColumn = w.width
	}
	w.cursor.row = newRow
	w.cursor.column = newColumn
}

func (w *Window) LoadBuffer(b Buffer) error {
	w.buffer = b
	return b.Load()
}

func (w *Window) Render() string {
	var sb strings.Builder
	io.Copy(&sb, w.buffer)
	return sb.String()
}
