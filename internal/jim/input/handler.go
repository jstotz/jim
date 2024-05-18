package input

import (
	"fmt"

	"github.com/jstotz/jim/internal/jim/command"
	"github.com/jstotz/jim/internal/jim/modes"
)

const (
	KeyEnter     = rune(13)
	KeyEscape    = rune(27)
	KeyBackspace = rune(127)
)

func HandleKeyPress(mode modes.Mode, c rune) (command.Command, error) {
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

func handleNormalKeyPress(c rune) (command.Command, error) {
	switch c {
	case 'i':
		return command.ActivateMode{Mode: modes.ModeInsert}, nil
	case ':':
		return command.ActivateMode{Mode: modes.ModeCommand}, nil
	case 'q':
		return command.Exit{}, nil
	case 'j':
		return command.MoveCursorRelative{DeltaRows: 1, DeltaColumns: 0}, nil
	case 'k':
		return command.MoveCursorRelative{DeltaRows: -1, DeltaColumns: 0}, nil
	case 'h', KeyBackspace:
		return command.MoveCursorRelative{DeltaRows: 0, DeltaColumns: -1}, nil
	case 'l':
		return command.MoveCursorRelative{DeltaRows: 0, DeltaColumns: 1}, nil
	case 'x':
		return command.DeleteText{Length: 1}, nil
	default:
		return command.Noop{}, nil
	}
}

func handleCommonEditModeKeyPress(c rune) (command.Command, error) {
	switch c {
	case KeyEscape:
		return command.ActivateMode{Mode: modes.ModeNormal}, nil
	case KeyBackspace:
		return command.DeleteText{Length: -1}, nil
	default:
		return command.InsertText{Text: string(c)}, nil
	}
}

func handleInsertKeyPress(c rune) (command.Command, error) {
	return handleCommonEditModeKeyPress(c)
}

func handleCommandKeyPress(c rune) (command.Command, error) {
	switch c {
	case KeyEnter:
		return command.EvalCommandBuffer{}, nil
	default:
		return handleCommonEditModeKeyPress(c)
	}
}
