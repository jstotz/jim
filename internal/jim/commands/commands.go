package commands

import "github.com/jstotz/jim/internal/jim/modes"

type Command interface {
	command()
}

// Noop does nothing. Used to satisfy the return type there is no action to take.
type Noop struct{}

func (Noop) command() {}

// ActivateMode switches the editor mode
type ActivateMode struct {
	Mode modes.Mode
}

func (ActivateMode) command() {}

// InsertText insert the given text in the active buffer
type InsertText struct {
	Text string
}

func (InsertText) command() {}

// DeleteText deletes the given length of text starting from the current cursor position forward
// or backward if length is negative
type DeleteText struct {
	Length int
}

func (DeleteText) command() {}

// MoveCursorRelative moves the cursor position up and down and/or forward or backward the given
// number of rows and/or columns
type MoveCursorRelative struct {
	DeltaRows    int
	DeltaColumns int
}

func (MoveCursorRelative) command() {}

// Exit signals the editor to shut down and exit the process
type Exit struct{}

func (Exit) command() {}

// EvalCommandBuffer evaluates the contents of the command buffer
type EvalCommandBuffer struct{}

func (EvalCommandBuffer) command() {}

// Save persists the currently active buffer
type Save struct{}

func (Save) command() {}
