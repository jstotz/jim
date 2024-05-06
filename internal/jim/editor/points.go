package editor

type Point struct {
	row    int
	column int
}

func (p Point) RowIndex() int {
	return p.row - 1
}

func (p Point) ColumnIndex() int {
	return p.column - 1
}
