package input

import (
	"fmt"

	"github.com/jstotz/jim/internal/jim/commands"
	"github.com/jstotz/jim/internal/jim/modes"
)

const (
	KeyEnter     = rune(13)
	KeyEscape    = rune(27)
	KeyBackspace = rune(127)
)

func HandleKeyPress(mode modes.Mode, c rune) (commands.Command, error) {
	switch mode {
	case modes.ModeNormal:
		return handleNormalKeyPress(c)
	case modes.ModeInsert:
		return handleInsertKeyPress(c)
	case modes.ModeCommand:
		return handleCommandKeyPress(c)
	default:
		return nil, fmt.Errorf("unknown mode: %v", mode)
	}
}

func handleNormalKeyPress(c rune) (commands.Command, error) {
	switch c {
	case 'i':
		return commands.ActivateMode{Mode: modes.ModeInsert}, nil
	case ':':
		return commands.ActivateMode{Mode: modes.ModeCommand}, nil
	case 'q':
		return commands.Exit{}, nil
	case 'j':
		return commands.MoveCursorRelative{DeltaRows: 1, DeltaColumns: 0}, nil
	case 'k':
		return commands.MoveCursorRelative{DeltaRows: -1, DeltaColumns: 0}, nil
	case 'h', KeyBackspace:
		return commands.MoveCursorRelative{DeltaRows: 0, DeltaColumns: -1}, nil
	case 'l':
		return commands.MoveCursorRelative{DeltaRows: 0, DeltaColumns: 1}, nil
	case 'x':
		return commands.DeleteText{Length: 1}, nil
	default:
		return commands.Noop{}, nil
	}
}

func handleCommonEditModeKeyPress(c rune) (commands.Command, error) {
	switch c {
	case KeyEscape:
		return commands.ActivateMode{Mode: modes.ModeNormal}, nil
	case KeyBackspace:
		return commands.DeleteText{Length: -1}, nil
	default:
		return commands.InsertText{Text: string(c)}, nil
	}
}

func handleInsertKeyPress(c rune) (commands.Command, error) {
	return handleCommonEditModeKeyPress(c)
}

func handleCommandKeyPress(c rune) (commands.Command, error) {
	switch c {
	case KeyEnter:
		return commands.EvalCommandBuffer{}, nil
	default:
		return handleCommonEditModeKeyPress(c)
	}
}
