package config

import "github.com/jstotz/jim/internal/jim/modes"

type Config struct {
	KeyBindings []KeyBinding
}

type KeyBinding struct {
	Mode   modes.Mode
	Keys   string
	Script string
}
