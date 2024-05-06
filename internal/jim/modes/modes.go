package modes

type Mode int

const (
	ModeNormal Mode = iota
	ModeInsert
	ModeCommand
)

// String - Creating common behavior - give the type a String function
func (m Mode) String() string {
	return [...]string{"Normal", "Insert", "Command"}[m]
}
