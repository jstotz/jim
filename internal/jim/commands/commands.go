package commands

import "github.com/jstotz/jim/internal/jim/modes"

type Command interface {
	command()
}
type Noop struct{}

func (Noop) command() {}

type ActivateMode struct {
	Mode modes.Mode
}

func (ActivateMode) command() {}

type InsertText struct {
	Text string
}

func (InsertText) command() {}

type DeleteText struct {
	Length int
}

func (DeleteText) command() {}

type MoveCursorRelative struct {
	DeltaRows    int
	DeltaColumns int
}

func (MoveCursorRelative) command() {}

type Exit struct{}

func (Exit) command() {}

type EvalCommandBuffer struct{}

func (EvalCommandBuffer) command() {}

type Save struct{}

func (Save) command() {}
