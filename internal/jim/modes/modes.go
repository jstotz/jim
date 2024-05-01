package modes

type Mode int

const (
	ModeNormal Mode = iota
	ModeInsert
	ModeCommand
)
