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
	"github.com/jstotz/jim/internal/jim/config"
	"github.com/jstotz/jim/internal/jim/input"
	"github.com/jstotz/jim/internal/jim/modes"
	"github.com/muesli/termenv"
	lua "github.com/yuin/gopher-lua"
	"golang.org/x/term"
)

const (
	cursorStyleBlock = "\033[2 q"
	cursorStyleLine  = "\033[5 q"
)

type Editor struct {
	mode          modes.Mode
	Logger        *slog.Logger
	tty           *os.File
	exitChan      chan error
	keypressChan  chan rune
	input         io.Reader
	output        *termenv.Output
	window        *Window
	prevTermState *term.State
	commandWindow *Window
	luaState      *lua.LState
	inputHandler  *input.Handler
}

func NewEditor(inputFile *os.File, outputFile *os.File, log *os.File) *Editor {
	var tty *os.File
	if inputFile == nil {
		inputFile = os.Stdin
		tty = inputFile
	}
	if outputFile == nil {
		outputFile = os.Stdout
	}

	logger := slog.New(slog.NewTextHandler(log, &slog.HandlerOptions{Level: slog.LevelDebug}))

	return &Editor{
		Logger:        logger,
		mode:          modes.ModeNormal,
		tty:           tty,
		exitChan:      make(chan error, 1),
		keypressChan:  make(chan rune, 1),
		input:         inputFile,
		output:        termenv.NewOutput(outputFile),
		prevTermState: nil,
		luaState:      lua.NewState(),
		// TODO: Allow config customization
		inputHandler: input.NewHandler(config.DefaultConfig()),
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
	e.window = NewWindow(nil, 0, 0, width, height-1, e.Logger)

	e.commandWindow = NewWindow(NewMemoryBuffer(e.Logger), height-1, 1, width, 1, e.Logger)

	return nil
}

func (e *Editor) LoadFile(path string) error {
	fb := NewFileBuffer(path, e.Logger)
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
	e.Logger.Info("Handling keypress", "key", c)
	cmd, err := e.inputHandler.HandleKeyPress(e.mode, c)
	if err != nil {
		return err
	}
	return e.runCommand(cmd)
}

func (e *Editor) parseCommand(expr string) (command.Command, error) {
	if strings.HasPrefix(expr, "lua ") {
		return command.EvalLua{Script: expr[4:]}, nil
	}
	if expr == "w" {
		return command.Save{}, nil
	}
	if expr == "q" {
		return command.Exit{}, nil
	}
	return command.Noop{}, fmt.Errorf("invalid expression: %s", expr)
}

func (e *Editor) evalCommand(expr string) error {
	cmd, err := e.parseCommand(expr)
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
	case command.EvalLua:
		return e.evalLua(cmd.Script)
	case command.Exit:
		e.exit(nil)
	default:
		return fmt.Errorf("unsupported command: %#v", cmd)
	}
	return nil
}

func (e *Editor) saveBuffer() error {
	written, err := e.window.buffer.Save()
	e.Logger.Debug("Saved buffer", "written", written)
	return err
}

func (e *Editor) evalCommandBuffer() error {
	expr := strings.TrimSpace(e.commandWindow.buffer.String())
	if err := e.evalCommand(expr); err != nil {
		e.Logger.Error("eval command buffer error", "err", err)
	}
	e.commandWindow.Clear()
	e.must(e.activateMode(modes.ModeNormal))
	return nil
}

func (e *Editor) evalLua(script string) error {
	e.Logger.Debug("running lua script", "script", script)
	if err := e.luaState.DoString(script); err != nil {
		e.Logger.Error("eval lua error", "err", err)
		return err
	}
	return nil
}

func (e *Editor) activateMode(mode modes.Mode) error {
	e.mode = mode
	e.commandWindow.Clear()
	e.Logger.Debug("Activated mode", "mode", mode)
	return nil
}

func (e *Editor) exit(err error) {
	e.exitChan <- err
}

func (e *Editor) Start() error {
	defer e.cleanup()
	e.output.AltScreen()

	NewAPIModule(e).Load()
	defer e.luaState.Close()

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
