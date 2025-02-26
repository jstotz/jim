package input

import (
	"github.com/jstotz/jim/internal/jim/command"
	"github.com/jstotz/jim/internal/jim/config"
	"github.com/jstotz/jim/internal/jim/modes"
)

type Handler struct {
	config config.Config
}

func NewHandler(cfg config.Config) *Handler {
	return &Handler{
		config: cfg,
	}
}

func (h *Handler) HandleKeyPress(mode modes.Mode, c rune) (command.Command, error) {
	keyStr := string(c)
	
	// Find matching key binding for the current mode and key
	for _, binding := range h.config.KeyBindings {
		if binding.Mode == mode && binding.Keys == keyStr {
			return binding.Command, nil
		}
	}

	// For insert and command modes, if no specific binding is found,
	// default to inserting the character
	if mode == modes.ModeInsert || mode == modes.ModeCommand {
		return command.InsertText{Text: keyStr}, nil
	}

	return command.Noop{}, nil
}
