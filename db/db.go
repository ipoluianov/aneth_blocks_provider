package db

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ipoluianov/aneth_blocks_provider/utils"
	"github.com/ipoluianov/gomisc/logger"
)

type DB struct {
	mtx sync.Mutex

	status    string
	substatus string

	// Settings
	network  string
	url      string
	periodMs int

	blockNumberDepth uint64
	timeDepth        uint64

	// Data
	latestBlockNumber uint64
	existingBlocks    map[uint64]struct{}
	blocksCache       map[uint64]*Block

	// Runtime
	client *ethclient.Client
}

var Instance *DB

func init() {
	Instance = NewDB("ETH", "https://eth.public-rpc.com", 2000)
}

func NewDB(network string, url string, periodMs int) *DB {
	var c DB
	c.network = network
	c.url = url
	c.periodMs = periodMs
	c.existingBlocks = make(map[uint64]struct{})
	c.blocksCache = make(map[uint64]*Block)
	c.status = "init"
	c.latestBlockNumber = 0

	c.timeDepth = uint64(86400)
	secondsPerBlock := uint64(12)
	c.blockNumberDepth = c.timeDepth / secondsPerBlock

	return &c
}

func (c *DB) Start() {
	c.status = "starting"
	var err error
	c.client, err = ethclient.Dial(c.url)
	if err != nil {
		logger.Println(err)
	}

	c.status = "getting latest block number"
	for c.latestBlockNumber == 0 {
		c.updateLatestBlockNumber()
		time.Sleep(1 * time.Second)
	}

	logger.Println("Latest block number received: ", c.latestBlockNumber)

	c.LoadExistingBlocks()

	c.status = "db started"
	c.substatus = ""

	c.updateLatestBlockNumber()
	go c.thLoad()
	go c.thUpdateLatestBlock()
}

func (c *DB) Stop() {
}

func getFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func (c *DB) LoadExistingBlocks() {
	c.status = "getting file list from file system"
	files, err := getFiles("data/" + c.network + "/")
	if err != nil {
		logger.Println("DB::LoadExistingBlocks error", err)
	}

	c.status = "getting file list"
	for i, fileName := range files {
		c.substatus = fileName + " (" + fmt.Sprint(i) + "/" + fmt.Sprint(len(files)) + ")"
		logger.Println("DB::LoadExistingBlocks", "file", fileName, " ", i, "/", len(files))
		var bl Block
		err = bl.Read(fileName)
		if err != nil {
			continue
		}
		c.mtx.Lock()
		c.blocksCache[bl.Number] = &bl
		c.mtx.Unlock()
	}
}

func (c *DB) updateLatestBlockNumber() error {
	logger.Println("DB::updateLatestBlockNumber", c.network)

	client, err := ethclient.Dial(c.url)
	if err != nil {
		logger.Println("DB::updateLatestBlockNumber Error:", err)
		return err
	}
	block, err := client.BlockByNumber(context.Background(), nil)
	if err != nil {
		logger.Println(c.network, "DB::updateLatestBlockNumber Error:", err)
		return err
	}
	c.mtx.Lock()
	blockNum := block.Header().Number.Uint64()
	blockNum -= 3
	c.latestBlockNumber = blockNum
	c.mtx.Unlock()
	logger.Println("DB::updateLatestBlockNumber result:", c.network, block.Header().Number.Int64(), "set:", blockNum)
	return nil
}

func (c *DB) GetState() (dbState *DbState) {
	c.mtx.Lock()
	dbState = &DbState{}
	dbState.CountOfBlocks = len(c.blocksCache)
	dbState.Status = c.status
	dbState.SubStatus = c.substatus
	blocksArray := make([]*Block, 0)
	for _, bl := range c.blocksCache {
		blocksArray = append(blocksArray, bl)
	}
	c.mtx.Unlock()

	sort.Slice(blocksArray, func(i, j int) bool {
		return blocksArray[i].Number < blocksArray[j].Number
	})

	if len(blocksArray) > 0 {
		var currentRange DbStateBlockRange

		currentBlockNumber := uint64(0)
		for _, bl := range blocksArray {
			if currentBlockNumber == 0 {
				currentRange.Number1 = bl.Number
				currentRange.DtStr1 = time.Unix(int64(bl.Time), 0).Format("2006-01-02 15:04:05")
				currentRange.Number2 = bl.Number
				currentRange.DtStr2 = time.Unix(int64(bl.Time), 0).Format("2006-01-02 15:04:05")
				currentRange.Count = 1
				currentBlockNumber = bl.Number
				continue
			}

			if bl.Number != currentBlockNumber+1 {
				currentRange.Count = int(currentRange.Number2 - currentRange.Number1 + 1)
				dbState.LoadedBlocks = append(dbState.LoadedBlocks, currentRange)
				currentRange = DbStateBlockRange{}
				currentBlockNumber = bl.Number

				currentRange.Number1 = bl.Number
				currentRange.DtStr1 = time.Unix(int64(bl.Time), 0).Format("2006-01-02 15:04:05")
				currentRange.Number2 = bl.Number
				currentRange.DtStr2 = time.Unix(int64(bl.Time), 0).Format("2006-01-02 15:04:05")
				currentRange.Count = 1
				continue
			}

			currentBlockNumber = bl.Number
			currentRange.Number2 = bl.Number
			currentRange.DtStr2 = time.Unix(int64(bl.Time), 0).Format("2006-01-02 15:04:05")
		}
		currentRange.Count = int(currentRange.Number2 - currentRange.Number1 + 1)
		dbState.LoadedBlocks = append(dbState.LoadedBlocks, currentRange)
	}

	loadedBlocksSortByCount := make([]DbStateBlockRange, len(dbState.LoadedBlocks))
	copy(loadedBlocksSortByCount, dbState.LoadedBlocks)
	sort.Slice(loadedBlocksSortByCount, func(i, j int) bool {
		return loadedBlocksSortByCount[i].Count < loadedBlocksSortByCount[j].Count
	})

	txs := c.GetData(0, 0xFFFFFFFFFFFFFFFF)
	dbState.LoadedBlocksTimeRange = "-"
	if len(txs) > 0 {
		dtBegin := time.Unix(int64(txs[0].BlDT), 0).Format("2006-01-02 15-04-05")
		dtEnd := time.Unix(int64(txs[len(txs)-1].BlDT), 0).Format("2006-01-02 15-04-05")
		dbState.LoadedBlocksTimeRange = dtBegin + " - " + dtEnd
	}

	return
}

func (c *DB) GetData(timeBegin uint64, timeEnd uint64) []*Tx {
	c.mtx.Lock()
	blocksArray := make([]*Block, 0)
	for _, bl := range c.blocksCache {
		if bl.Time >= timeBegin && bl.Time <= timeEnd {
			blocksArray = append(blocksArray, bl)
		}
	}
	c.mtx.Unlock()

	sort.Slice(blocksArray, func(i, j int) bool {
		return blocksArray[i].Number < blocksArray[j].Number
	})

	loadedBlocks := make([]DbStateBlockRange, 0)
	if len(blocksArray) > 0 {
		var currentRange DbStateBlockRange

		currentBlockNumber := uint64(0)
		for _, bl := range blocksArray {
			if currentBlockNumber == 0 {
				currentRange.Number1 = bl.Number
				currentRange.DtStr1 = time.Unix(int64(bl.Time), 0).Format("2006-01-02 15:04:05")
				currentRange.Number2 = bl.Number
				currentRange.DtStr2 = time.Unix(int64(bl.Time), 0).Format("2006-01-02 15:04:05")
				currentRange.Count = 1
				currentBlockNumber = bl.Number
				continue
			}

			if bl.Number != currentBlockNumber+1 {
				currentRange.Count = int(currentRange.Number2 - currentRange.Number1 + 1)
				loadedBlocks = append(loadedBlocks, currentRange)
				currentRange = DbStateBlockRange{}
				currentBlockNumber = bl.Number

				currentRange.Number1 = bl.Number
				currentRange.DtStr1 = time.Unix(int64(bl.Time), 0).Format("2006-01-02 15:04:05")
				currentRange.Number2 = bl.Number
				currentRange.DtStr2 = time.Unix(int64(bl.Time), 0).Format("2006-01-02 15:04:05")
				currentRange.Count = 1
				continue
			}

			currentBlockNumber = bl.Number
			currentRange.Number2 = bl.Number
			currentRange.DtStr2 = time.Unix(int64(bl.Time), 0).Format("2006-01-02 15:04:05")
		}
		currentRange.Count = int(currentRange.Number2 - currentRange.Number1 + 1)
		loadedBlocks = append(loadedBlocks, currentRange)
	}

	loadedBlocksSortByCount := make([]DbStateBlockRange, len(loadedBlocks))
	copy(loadedBlocksSortByCount, loadedBlocks)
	sort.Slice(loadedBlocksSortByCount, func(i, j int) bool {
		return loadedBlocksSortByCount[i].Count < loadedBlocksSortByCount[j].Count
	})

	if len(loadedBlocksSortByCount) == 0 {
		return nil
	}

	biggestRange := loadedBlocksSortByCount[len(loadedBlocksSortByCount)-1]

	txs := make([]*Tx, 0)

	c.mtx.Lock()
	for blNumber := biggestRange.Number1; blNumber <= biggestRange.Number2; blNumber++ {
		bl, ok := c.blocksCache[blNumber]
		if ok && bl != nil {
			txs = append(txs, bl.Txs...)
		}
	}
	c.mtx.Unlock()

	sort.Slice(txs, func(i, j int) bool {
		return txs[i].BlDT < txs[j].BlDT
	})
	return txs
}

func (c *DB) GroupByMinutes(beginDT uint64, endDT uint64, txs []*Tx) *TxsByMinutes {
	logger.Println("An::GroupByMinutes begin")
	var res TxsByMinutes
	firstTxDt := beginDT
	lastTxDt := endDT

	firstTxDt = firstTxDt / 60
	firstTxDt = firstTxDt * 60

	lastTxDt = lastTxDt / 60
	lastTxDt = lastTxDt * 60

	countOfRanges := (lastTxDt - firstTxDt) / 60
	res.Items = make([]*TxsByMinute, countOfRanges)

	index := 0
	for dt := firstTxDt; dt < lastTxDt; dt += 60 {
		res.Items[index] = &TxsByMinute{}
		res.Items[index].DT = dt
		index++
	}

	for i := 0; i < len(txs); i++ {
		t := txs[i]
		rangeIndex := (t.BlDT - firstTxDt) / 60
		if int(rangeIndex) >= len(res.Items) {
			fmt.Println("OVERFLOW")
		}
		res.Items[rangeIndex].TXS = append(res.Items[rangeIndex].TXS, t)
	}

	logger.Println("An::GroupByMinutes end")

	return &res
}

func (c *DB) GetLatestTransactions() (*TxsByMinutes, []*Tx) {
	lastSeconds := uint64(24 * 3600)
	lastTxDt := uint64(time.Now().UTC().Unix())
	firstTxDt := uint64(lastTxDt - lastSeconds)
	firstTxDt = firstTxDt / 60
	firstTxDt = firstTxDt * 60
	lastTxDt = lastTxDt / 60
	lastTxDt = (lastTxDt + 1) * 60
	txs := c.GetData(firstTxDt, lastTxDt)
	if len(txs) < 1 {
		return &TxsByMinutes{}, nil
	}
	byMinutes := c.GroupByMinutes(firstTxDt, lastTxDt, txs)
	return byMinutes, txs
}

func (c *DB) LatestBlockNumber() uint64 {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	return c.latestBlockNumber
}

func (c *DB) BlockExists(blockNumber uint64) bool {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	if _, ok := c.existingBlocks[blockNumber]; ok {
		return true
	}
	//dir := c.blockDir(blockNumber)
	fileName := c.blockFile(blockNumber)
	st, err := os.Stat(fileName)
	if err != nil {
		return false
	}
	if st.IsDir() {
		return false
	}
	c.existingBlocks[blockNumber] = struct{}{}
	return true
}

func (c *DB) firstBlockToLoad() uint64 {
	return c.latestBlockNumber - c.blockNumberDepth
}

func (c *DB) loadNextBlock() {
	blockNumberToLoad := uint64(0)
	for blockNumber := c.latestBlockNumber; blockNumber > 0; blockNumber-- {
		if !c.BlockExists(blockNumber) {
			blockNumberToLoad = blockNumber
			break
		}
	}

	if blockNumberToLoad < c.firstBlockToLoad()+10 {
		logger.Println("DB::loadNextBlock", "no block to load:", blockNumberToLoad, "latest block:", c.latestBlockNumber)
		return
	}

	logger.Println("DB::loadNextBlock", c.network, "Getting Block:", blockNumberToLoad)
	block, err := c.client.BlockByNumber(context.Background(), big.NewInt(int64(blockNumberToLoad)))
	if err != nil {
		logger.Println(c.network, "Getting Latest Block Error:", err)
		return
	}

	var b Block
	b.Number = blockNumberToLoad
	b.Time = block.Header().Time

	for _, t := range block.Transactions() {
		var tx Tx
		tx.BlNumber = uint64(b.Number)
		tx.BlDT = b.Time
		tx.TxFrom = utils.TrFrom(t)
		tx.TxTo = t.To()
		tx.TxData = t.Data()
		tx.TxValue = t.Value()
		b.Txs = append(b.Txs, &tx)
	}

	c.SaveBlock(&b)
	c.mtx.Lock()
	c.blocksCache[b.Number] = &b
	c.mtx.Unlock()
}

func (c *DB) normilizeBlockNumberString(blockNumber uint64) string {
	blockNumberString := fmt.Sprint(blockNumber)
	for len(blockNumberString) < 12 {
		blockNumberString = "0" + blockNumberString
	}
	result := make([]byte, 0)
	for i := 0; i < len(blockNumberString); i++ {
		if (i%3) == 0 && i > 0 {
			result = append(result, '-')
		}
		result = append(result, blockNumberString[i])
	}
	return string(result)
}

func (c *DB) blockDir(blockNumber uint64) string {
	dir := "data/" + c.network + "/" + c.normilizeBlockNumberString(blockNumber-(blockNumber%10000))
	return dir
}

func (c *DB) blockFile(blockNumber uint64) string {
	fileName := c.blockDir(blockNumber) + "/" + c.normilizeBlockNumberString(blockNumber) + ".block"
	return fileName
}

func (c *DB) SaveBlock(b *Block) error {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	dir := c.blockDir(b.Number)
	err := os.MkdirAll(dir, 0777)
	if err != nil {
		logger.Println(c.network, "write block error:", err)
		return err
	}
	fileName := c.blockFile(b.Number)
	return b.Write(fileName)
}

func (c *DB) GetBlock(blockNumber uint64) (*Block, error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	var b *Block
	var err error
	if bl, ok := c.blocksCache[blockNumber]; ok {
		b = bl
	} else {
		b = &Block{}
		fileName := c.blockFile(blockNumber)
		err = b.Read(fileName)
		if err == nil {
			c.blocksCache[blockNumber] = b
		}
	}
	return b, err
}

func (c *DB) GetBlockFromCache(blockNumber uint64) (*Block, error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	var b *Block
	var err error
	if bl, ok := c.blocksCache[blockNumber]; ok {
		b = bl
	} else {
		err = errors.New("not found")
	}
	return b, err
}

func (c *DB) thLoad() {
	logger.Println("DB::ThUpdate begin", c.network)

	for {
		c.loadNextBlock()
		time.Sleep(time.Duration(c.periodMs) * time.Millisecond)
	}
}

func (c *DB) thUpdateLatestBlock() {
	logger.Println("DB::thUpdateLatestBlock", c.network)

	for {
		c.updateLatestBlockNumber()
		c.removeOldBlocks()
		time.Sleep(5000 * time.Millisecond)
	}
}

func (c *DB) removeOldBlocks() {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	leftBorderOfTime := uint64(time.Now().Unix()) - c.timeDepth

	blocksToRemove := make([]uint64, 0)
	for _, bl := range c.blocksCache {
		if bl.Time < leftBorderOfTime {
			blocksToRemove = append(blocksToRemove, bl.Number)
		}
	}

	for _, bNum := range blocksToRemove {
		delete(c.blocksCache, bNum)
		filePath := c.blockFile(bNum)
		err := os.Remove(filePath)
		if err != nil {
			logger.Println("DB::removeOldBlocks", "remove file error:", err)
		} else {
			logger.Println("DB::removeOldBlocks", "remove file", filePath, "SUCCESS")
		}
	}
}
