package db

type DbStateBlockRange struct {
	DtStr1  string
	Number1 uint64
	DtStr2  string
	Number2 uint64
	Count   int
}

type DbState struct {
	MinBlock              int64
	MaxBlock              int64
	CountOfBlocks         int
	Network               string
	Status                string
	SubStatus             string
	LoadedBlocks          []DbStateBlockRange
	LoadedBlocksTimeRange string
}
