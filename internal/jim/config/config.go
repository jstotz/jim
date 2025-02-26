package config

import (
	"github.com/jstotz/jim/internal/jim/command"
	"github.com/jstotz/jim/internal/jim/modes"
)

const (
	KeyEnter     = rune(13)
	KeyEscape    = rune(27)
	KeyBackspace = rune(127)
)

type Config struct {
	KeyBindings []KeyBinding
}

type KeyBinding struct {
	Mode    modes.Mode
	Keys    string
	Command command.Command
}
