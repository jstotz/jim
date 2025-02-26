package config

import "github.com/jstotz/jim/internal/jim/modes"

func DefaultConfig() Config {
	return Config{
		KeyBindings: []KeyBinding{
			{
				Mode:   modes.ModeNormal,
				Keys:   "x",
				Script: "jim.api.delete()",
			},
		},
	}
}
