package an

type ResultItemByMinutes struct {
	Index int
	DT    uint64
	DTStr string
	Value float64
}

type ResultItemString struct {
	Text string
}

type Result struct {
	Code            string
	Count           int
	CurrentDateTime string
	ItemsByMinutes  []*ResultItemByMinutes
	ItemsString     []ResultItemString
}
