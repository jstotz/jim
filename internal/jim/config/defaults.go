package config

import (
	"github.com/jstotz/jim/internal/jim/command"
	"github.com/jstotz/jim/internal/jim/modes"
)

func DefaultConfig() Config {
	return Config{
		KeyBindings: []KeyBinding{
			// Normal mode bindings
			{
				Mode:    modes.ModeNormal,
				Keys:    "i",
				Command: command.ActivateMode{Mode: modes.ModeInsert},
			},
			{
				Mode:    modes.ModeNormal,
				Keys:    ":",
				Command: command.ActivateMode{Mode: modes.ModeCommand},
			},
			{
				Mode:    modes.ModeNormal,
				Keys:    "q",
				Command: command.Exit{},
			},
			{
				Mode:    modes.ModeNormal,
				Keys:    "j",
				Command: command.MoveCursorRelative{DeltaRows: 1, DeltaColumns: 0},
			},
			{
				Mode:    modes.ModeNormal,
				Keys:    "k",
				Command: command.MoveCursorRelative{DeltaRows: -1, DeltaColumns: 0},
			},
			{
				Mode:    modes.ModeNormal,
				Keys:    "h",
				Command: command.MoveCursorRelative{DeltaRows: 0, DeltaColumns: -1},
			},
			{
				Mode:    modes.ModeNormal,
				Keys:    string(KeyBackspace),
				Command: command.MoveCursorRelative{DeltaRows: 0, DeltaColumns: -1},
			},
			{
				Mode:    modes.ModeNormal,
				Keys:    "l",
				Command: command.MoveCursorRelative{DeltaRows: 0, DeltaColumns: 1},
			},
			{
				Mode:    modes.ModeNormal,
				Keys:    "x",
				Command: command.DeleteText{Length: 1},
			},
			// Insert mode bindings
			{
				Mode:    modes.ModeInsert,
				Keys:    string(KeyEscape),
				Command: command.ActivateMode{Mode: modes.ModeNormal},
			},
			{
				Mode:    modes.ModeInsert,
				Keys:    string(KeyBackspace),
				Command: command.DeleteText{Length: -1},
			},
			// Command mode bindings
			{
				Mode:    modes.ModeCommand,
				Keys:    string(KeyEnter),
				Command: command.EvalCommandBuffer{},
			},
			{
				Mode:    modes.ModeCommand,
				Keys:    string(KeyEscape),
				Command: command.ActivateMode{Mode: modes.ModeNormal},
			},
			{
				Mode:    modes.ModeCommand,
				Keys:    string(KeyBackspace),
				Command: command.DeleteText{Length: -1},
			},
		},
	}
}
