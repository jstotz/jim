package editor

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"strings"

	"github.com/jstotz/jim/internal/jim/command"
	"github.com/jstotz/jim/internal/jim/input"
	"github.com/jstotz/jim/internal/jim/modes"
	"github.com/muesli/termenv"
	"golang.org/x/term"
)

const (
	cursorStyleBlock = "\033[2 q"
	cursorStyleLine  = "\033[5 q"
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
	commandWindow *Window
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
		mode:          modes.ModeNormal,
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

func (e *Editor) FocusedWindow() *Window {
	if e.mode == modes.ModeCommand {
		return e.commandWindow
	}
	return e.window
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
	e.window = NewWindow(nil, 0, 0, width, height-1, e.logger)

	e.commandWindow = NewWindow(NewMemoryBuffer(e.logger), height-1, 1, width, 1, e.logger)

	return nil
}

func (e *Editor) LoadFile(path string) error {
	fb := NewFileBuffer(path, e.logger)
	return e.FocusedWindow().LoadBuffer(fb)
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

func (e *Editor) updateCursor() {
	w := e.FocusedWindow()
	e.output.MoveCursor(w.cursor.row+w.rowOffset, w.cursor.column+w.columnOffset)
	e.setCursorStyle()
}

func (e *Editor) setCursorStyle() {
	switch e.mode {
	case modes.ModeNormal:
		e.mustWriteString(cursorStyleBlock)
	case modes.ModeInsert, modes.ModeCommand:
		e.mustWriteString(cursorStyleLine)
	}
}

func (e *Editor) mustWriteString(s string) (written int) {
	written, err := e.output.WriteString(s)
	if err != nil {
		e.exit(err)
		return 0
	}
	return written
}

func (e *Editor) handleKeypress(c rune) error {
	e.logger.Info("Handling keypress", "key", c)
	cmd, err := input.HandleKeyPress(e.mode, c)
	if err != nil {
		return err
	}
	return e.runCommand(cmd)
}

func (e *Editor) parseExpr(expr string) (command.Command, error) {
	if expr == "w" {
		return command.Save{}, nil
	}
	if expr == "q" {
		return command.Exit{}, nil
	}
	return command.Noop{}, fmt.Errorf("invalid expression: %s", expr)
}

func (e *Editor) eval(expr string) error {
	cmd, err := e.parseExpr(expr)
	if err != nil {
		return err
	}
	return e.runCommand(cmd)
}

func (e *Editor) runCommand(cmd command.Command) error {
	w := e.FocusedWindow()
	switch cmd := cmd.(type) {
	case command.Noop:
		return nil
	case command.Save:
		return e.saveBuffer()
	case command.MoveCursorRelative:
		e.FocusedWindow().MoveCursorRelative(cmd.DeltaRows, cmd.DeltaColumns)
	case command.DeleteText:
		return w.DeleteText(w.CurrentPosition(), cmd.Length)
	case command.InsertText:
		return w.InsertText(w.CurrentPosition(), cmd.Text)
	case command.ActivateMode:
		return e.activateMode(cmd.Mode)
	case command.EvalCommandBuffer:
		return e.evalCommandBuffer()
	case command.Exit:
		e.exit(nil)
	default:
		return fmt.Errorf("unsupported command: %#v", cmd)
	}
	return nil
}

func (e *Editor) saveBuffer() error {
	written, err := e.window.buffer.Save()
	e.logger.Debug("Saved buffer", "written", written)
	return err
}

func (e *Editor) evalCommandBuffer() error {
	expr := strings.TrimSpace(e.commandWindow.buffer.String())
	if err := e.eval(expr); err != nil {
		e.logger.Debug("eval command buffer", "err", err)
	}
	e.commandWindow.Clear()
	e.must(e.activateMode(modes.ModeNormal))
	return nil
}

func (e *Editor) activateMode(mode modes.Mode) error {
	e.mode = mode
	e.commandWindow.Clear()
	e.logger.Debug("Activated mode", "mode", mode)
	return nil
}

func (e *Editor) exit(err error) {
	e.exitChan <- err
}

func (e *Editor) Start() error {
	defer e.cleanup()
	e.output.AltScreen()
	go e.readInput()
	e.must(e.draw())
	e.updateCursor()
	for {
		select {
		case c := <-e.keypressChan:
			e.must(e.handleKeypress(c))
			e.output.ClearScreen()
			e.must(e.draw())
			e.updateCursor()
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

func (e *Editor) draw() error {
	_, err := e.output.WriteString(strings.Join([]string{
		e.window.Render(),
		e.renderStatusLine(),
	}, "\r\n"))
	return err
}

func (e *Editor) renderStatusLine() string {
	if e.mode == modes.ModeCommand {
		content := e.commandWindow.Render()
		return fmt.Sprintf(":%s", content)
	}
	return fmt.Sprintf("[%s]", strings.ToUpper(e.mode.String()))
}

func (e *Editor) cleanup() {
	e.output.ExitAltScreen()
	if err := term.Restore(int(e.tty.Fd()), e.prevTermState); err != nil {
		log.Printf("error restoring terminal state: %v", err)
	}
}
