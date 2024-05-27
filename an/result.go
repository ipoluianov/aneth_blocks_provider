package an

type ResultItem struct {
	Index int
	DT    uint64
	DTStr string
	Value float64
}

type Result struct {
	Count           int
	CurrentDateTime string
	Items           []*ResultItem
}
