package jed

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/muesli/termenv"
	"golang.org/x/term"
)

type Point struct {
	row    int64
	column int64
}

type Editor struct {
	tty            *os.File
	exitChan       chan error
	keypressChan   chan rune
	input          io.Reader
	output         *termenv.Output
	activeBuffer   Buffer
	cursorPosition Point
	prevTermState  *term.State
}

func NewEditor(input *os.File, output *os.File) *Editor {
	var tty *os.File
	if input == nil {
		input = os.Stdin
		tty = input
	}
	if output == nil {
		output = os.Stdout
	}
	return &Editor{
		tty:            tty,
		exitChan:       make(chan error),
		keypressChan:   make(chan rune),
		input:          input,
		output:         termenv.NewOutput(output),
		activeBuffer:   nil,
		prevTermState:  nil,
		cursorPosition: Point{row: 0, column: 0},
	}
}

func (e *Editor) Setup() error {
	prevTermState, err := term.MakeRaw(int(e.tty.Fd()))
	e.prevTermState = prevTermState
	if err != nil {
		return fmt.Errorf("terminal raw mode: %w", err)
	}
	return nil
}

func (e *Editor) LoadFile(path string) error {
	fb := NewFileBuffer(path)
	e.activeBuffer = fb
	return fb.Load()
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

func (e *Editor) moveCursorRelative(deltaRow int64, deltaColumn int64) {
	// TODO: bounds checking
	e.cursorPosition.row += deltaRow
	e.cursorPosition.column += deltaColumn
}

func (e *Editor) handleKeypress(c rune) {
	switch c {
	case 'q':
		e.exitChan <- nil
	case 'j':
		e.moveCursorRelative(-1, 0)
	case 'k':
		e.moveCursorRelative(1, 0)
	case 'h':
		e.moveCursorRelative(0, -1)
	case 'l':
		e.moveCursorRelative(0, 1)
	}
}

func (e *Editor) Start() error {
	defer e.cleanup()
	e.output.AltScreen()
	go e.readInput()
	e.render()
	for {
		select {
		case c := <-e.keypressChan:
			e.handleKeypress(c)
			e.render()
		case err := <-e.exitChan:
			return err
		}
	}
}

func (e *Editor) render() error {
	_, err := io.Copy(e.output, e.activeBuffer)
	return err
}

func (e *Editor) cleanup() {
	e.output.ExitAltScreen()
	if err := term.Restore(int(e.tty.Fd()), e.prevTermState); err != nil {
		log.Printf("error restoring terminal state: %v", err)
	}
}
