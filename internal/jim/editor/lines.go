package editor

type Line struct {
	number  int64
	content string
}

type LineRange struct {
	start int64
	end   int64
}

func (lr LineRange) ShiftBy(n int64) LineRange {
	return LineRange{start: lr.start + n, end: lr.end + n}
}
