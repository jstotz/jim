package input

import (
	"fmt"

	"github.com/jstotz/jim/internal/jim/commands"
	"github.com/jstotz/jim/internal/jim/modes"
)

func HandleKeyPress(mode modes.Mode, c rune) (commands.Command, error) {
	switch mode {
	case modes.ModeNormal:
		return handleNormalKeyPress(c)
	case modes.ModeInsert:
		panic("insert mode not implemented")
	case modes.ModeCommand:
		panic("command mode not implemented")
	}
	return nil, fmt.Errorf("unknown mode: %v", mode)
}

func handleNormalKeyPress(c rune) (commands.Command, error) {
	switch c {
	case 'q':
		return commands.Exit{}, nil
	case 'j':
		return commands.MoveCursorRelative{DeltaRows: 1, DeltaColumns: 0}, nil
	case 'k':
		return commands.MoveCursorRelative{DeltaRows: -1, DeltaColumns: 0}, nil
	case 'h':
		return commands.MoveCursorRelative{DeltaRows: 0, DeltaColumns: -1}, nil
	case 'l':
		return commands.MoveCursorRelative{DeltaRows: 0, DeltaColumns: 1}, nil
	}
	return commands.Noop{}, nil
}
