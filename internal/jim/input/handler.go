package input

import (
	"fmt"

	"github.com/jstotz/jim/internal/jim/commands"
	"github.com/jstotz/jim/internal/jim/modes"
)

const KeyEscape = rune(27)

func HandleKeyPress(mode modes.Mode, c rune) (commands.Command, error) {
	switch mode {
	case modes.ModeNormal:
		return handleNormalKeyPress(c)
	case modes.ModeInsert:
		return handleInsertKeyPress(c)
	case modes.ModeCommand:
		panic("command mode not implemented")
	default:
		return nil, fmt.Errorf("unknown mode: %v", mode)
	}
}

func handleNormalKeyPress(c rune) (commands.Command, error) {
	switch c {
	case 'i':
		return commands.ActivateMode{Mode: modes.ModeInsert}, nil
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
	default:
		return commands.Noop{}, nil
	}
}

func handleInsertKeyPress(c rune) (commands.Command, error) {
	switch c {
	case KeyEscape:
		return commands.ActivateMode{Mode: modes.ModeNormal}, nil
	default:
		return commands.InsertText{Text: string(c)}, nil
	}
}
