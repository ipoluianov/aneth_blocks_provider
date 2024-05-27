package db

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ipoluianov/gomisc/logger"
)

type DB struct {
	mtx sync.Mutex

	// Settings
	network  string
	url      string
	periodMs int

	// Data
	latestBlockNumber int64
	existingBlocks    map[int64]struct{}
	blocksCache       map[int64]*Block

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
	c.existingBlocks = make(map[int64]struct{})
	c.blocksCache = make(map[int64]*Block)
	return &c
}

func (c *DB) Start() {
	var err error
	c.client, err = ethclient.Dial(c.url)
	if err != nil {
		log.Println(err)
	}

	c.updateLatestBlockNumber()
	go c.thLoad()
	go c.thUpdateLatestBlock()
}

func (c *DB) Stop() {
}

func (c *DB) updateLatestBlockNumber() error {
	log.Println(c.network, "UpdateLatestBlockNumber")

	client, err := ethclient.Dial(c.url)
	if err != nil {
		log.Println("UpdateLatestBlockNumber Error:", err)
		return err
	}
	block, err := client.BlockByNumber(context.Background(), nil)
	if err != nil {
		log.Println(c.network, "UpdateLatestBlockNumber Error:", err)
		return err
	}
	c.mtx.Lock()
	blockNum := block.Header().Number.Int64()
	blockNum -= 3
	c.latestBlockNumber = blockNum
	c.mtx.Unlock()
	log.Println(c.network, "UpdateLatestBlockNumber result:", block.Header().Number.Int64(), "set:", blockNum)
	return nil
}

func (c *DB) State() (minBlock int64, maxBlock int64, countOfBlocks int, network string) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	return
}

func (c *DB) LatestBlockNumber() int64 {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	return c.latestBlockNumber
}

func (c *DB) BlockExists(blockNumber int64) bool {
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

func (c *DB) loadNextBlock() {
	blockNumberToLoad := int64(-1)
	for blockNumber := c.latestBlockNumber; blockNumber > 0; blockNumber-- {
		if !c.BlockExists(blockNumber) {
			blockNumberToLoad = blockNumber
			break
		}
	}

	log.Println(c.network, "Getting Block:", blockNumberToLoad)
	block, err := c.client.BlockByNumber(context.Background(), big.NewInt(blockNumberToLoad))
	if err != nil {
		log.Println(c.network, "Getting Latest Block Error:", err)
		return
	}

	var b Block
	b.Number = blockNumberToLoad
	b.Header = *block.Header()
	b.Transactions = block.Transactions()

	c.SaveBlock(&b)
}

func (c *DB) normilizeBlockNumberString(blockNumber int64) string {
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

func (c *DB) blockDir(blockNumber int64) string {
	dir := "data/" + c.network + "/" + c.normilizeBlockNumberString(blockNumber-(blockNumber%10000))
	return dir
}

func (c *DB) blockFile(blockNumber int64) string {
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

func (c *DB) GetBlock(blockNumber int64) (*Block, error) {
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

func (c *DB) thLoad() {
	log.Println(c.network, "DB::ThUpdate begin")

	for {
		c.loadNextBlock()
		time.Sleep(time.Duration(c.periodMs) * time.Millisecond)
	}
}

func (c *DB) thUpdateLatestBlock() {
	log.Println(c.network, "DB::ThUpdate begin")

	for {
		c.updateLatestBlockNumber()
		time.Sleep(5000 * time.Millisecond)
	}
}
