package db

type DbState struct {
	MinBlock      int64
	MaxBlock      int64
	CountOfBlocks int
	Network       string
	Status        string
	SubStatus     string
}
