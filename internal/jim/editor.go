package jim

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"

	"github.com/muesli/termenv"
	"golang.org/x/term"
)

type Point struct {
	row    int
	column int
}

type Line struct {
	number  int64
	content string
}

type LineRange struct {
	start int64
	end   int64
}

func (lr LineRange) ShiftBy(n int64) LineRange {
	return LineRange{start: lr.start + n, end: lr.end + n}
}

type Editor struct {
	logger        *slog.Logger
	tty           *os.File
	exitChan      chan error
	keypressChan  chan rune
	input         io.Reader
	output        *termenv.Output
	window        *Window
	prevTermState *term.State
}

func NewEditor(input *os.File, output *os.File, log *os.File) *Editor {
	var tty *os.File
	if input == nil {
		input = os.Stdin
		tty = input
	}
	if output == nil {
		output = os.Stdout
	}

	logger := slog.New(slog.NewTextHandler(log, &slog.HandlerOptions{Level: slog.LevelDebug}))

	return &Editor{
		logger:        logger,
		tty:           tty,
		exitChan:      make(chan error),
		keypressChan:  make(chan rune),
		input:         input,
		output:        termenv.NewOutput(output),
		prevTermState: nil,
	}
}

func (e *Editor) GetSize() (width, height int, err error) {
	return term.GetSize(int(e.tty.Fd()))
}

func (e *Editor) Setup() error {
	prevTermState, err := term.MakeRaw(int(e.tty.Fd()))
	e.prevTermState = prevTermState
	if err != nil {
		return fmt.Errorf("terminal raw mode: %w", err)
	}
	width, height, err := e.GetSize()
	if err != nil {
		return err
	}
	e.window = NewWindow(width, height, e.logger)
	return nil
}

func (e *Editor) LoadFile(path string) error {
	fb := NewFileBuffer(path)
	return e.window.LoadBuffer(fb)
}

func (e *Editor) readInput() error {
	r := bufio.NewReader(e.input)
	for {
		c, _, err := r.ReadRune()
		if err != nil {
			e.exitChan <- err
			return err
		}
		e.keypressChan <- c
	}
}

func (e *Editor) updateCursorPosition() {
	e.output.MoveCursor(e.window.cursor.row, e.window.cursor.column)
}

func (e *Editor) handleKeypress(c rune) {
	w := e.window
	switch c {
	case 'q':
		close(e.exitChan)
	case 'j':
		w.MoveCursorRelative(1, 0)
	case 'k':
		w.MoveCursorRelative(-1, 0)
	case 'h':
		w.MoveCursorRelative(0, -1)
	case 'l':
		w.MoveCursorRelative(0, 1)
	}
}

func (e *Editor) Start() error {
	defer e.cleanup()
	e.output.AltScreen()
	go e.readInput()
	e.render()
	e.updateCursorPosition()
	for {
		select {
		case c := <-e.keypressChan:
			e.handleKeypress(c)
			e.output.ClearScreen()
			e.render()
			e.updateCursorPosition()
		case err := <-e.exitChan:
			return err
		}
	}
}

func (e *Editor) render() error {
	_, err := e.output.WriteString(e.window.Render())
	return err
}

func (e *Editor) cleanup() {
	e.output.ExitAltScreen()
	if err := term.Restore(int(e.tty.Fd()), e.prevTermState); err != nil {
		log.Printf("error restoring terminal state: %v", err)
	}
}
