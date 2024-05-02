package editor

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"

	"github.com/jstotz/jim/internal/jim/commands"
	"github.com/jstotz/jim/internal/jim/input"
	"github.com/jstotz/jim/internal/jim/modes"
	"github.com/muesli/termenv"
	"golang.org/x/term"
)

type Editor struct {
	mode          modes.Mode
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
		exitChan:      make(chan error, 1),
		keypressChan:  make(chan rune, 1),
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

func (e *Editor) readInput() {
	r := bufio.NewReader(e.input)
	for {
		c, _, err := r.ReadRune()
		if err != nil {
			e.exit(err)
			return
		}
		e.keypressChan <- c
	}
}

func (e *Editor) updateCursorPosition() {
	e.output.MoveCursor(e.window.cursor.row, e.window.cursor.column)
}

func (e *Editor) handleKeypress(c rune) error {
	cmd, err := input.HandleKeyPress(e.mode, c)
	if err != nil {
		return err
	}
	return e.runCommand(cmd)
}

func (e *Editor) runCommand(cmd commands.Command) error {
	switch cmd := cmd.(type) {
	case commands.Noop:
		return nil
	case commands.MoveCursorRelative:
		e.window.MoveCursorRelative(cmd.DeltaRows, cmd.DeltaColumns)
	case commands.InsertText:
		e.logger.Warn("i don't know how to insert text yet", "cmd", cmd)
	case commands.Exit:
		e.exit(nil)
		return nil
	}
	return nil
}

func (e *Editor) exit(err error) {
	e.exitChan <- err
}

func (e *Editor) Start() error {
	defer e.cleanup()
	e.output.AltScreen()
	go e.readInput()
	e.must(e.render())
	e.updateCursorPosition()
	for {
		select {
		case c := <-e.keypressChan:
			e.must(e.handleKeypress(c))
			e.output.ClearScreen()
			e.must(e.render())
			e.updateCursorPosition()
		case err := <-e.exitChan:
			return err
		}
	}
}

func (e *Editor) must(err error) {
	if err != nil {
		e.exit(err)
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
