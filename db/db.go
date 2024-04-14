package db

import (
	"context"
	"errors"
	"log"
	"math/big"
	"sort"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
)

type DB struct {
	mtx sync.Mutex

	// Settings
	network        string
	url            string
	periodMs       int
	maxBlocksCount int64

	// Data
	latestBlockNumber int64
	blocksMap         map[int64]*Block
	blocksList        []*Block

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
	c.blocksMap = make(map[int64]*Block)
	c.blocksList = make([]*Block, 0)
	c.maxBlocksCount = 10000
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
	blockNum -= 10
	c.latestBlockNumber = blockNum
	c.mtx.Unlock()
	log.Println(c.network, "UpdateLatestBlockNumber result:", block.Header().Number.Int64(), "set:", blockNum)
	return nil
}

func (c *DB) State() (minBlock int64, maxBlock int64, countOfBlocks int, network string) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	if len(c.blocksList) > 0 {
		minBlock = c.blocksList[0].Number
		maxBlock = c.blocksList[len(c.blocksList)-1].Number
	}
	countOfBlocks = len(c.blocksList)
	network = c.network
	return
}

func (c *DB) LatestBlockNumber() int64 {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	return c.latestBlockNumber
}

func (c *DB) GetBlock(blockNumber int64) (*Block, error) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	if b, ok := c.blocksMap[blockNumber]; ok {
		return b, nil
	}
	return nil, errors.New("not found")
}

func (c *DB) IsBlockExists(blockNumber int64) bool {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	if _, ok := c.blocksMap[blockNumber]; ok {
		return true
	}
	return false
}

func (c *DB) loadNextBlock() {
	blockNumberToLoad := int64(-1)
	for blockNumber := c.latestBlockNumber; blockNumber > 0; blockNumber-- {
		if !c.IsBlockExists(blockNumber) {
			blockNumberToLoad = blockNumber
			break
		}
	}

	if blockNumberToLoad < c.latestBlockNumber-int64(c.maxBlocksCount) {
		return
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
	c.addBlock(&b)
}

func (c *DB) addBlock(b *Block) error {
	c.mtx.Lock()
	c.blocksMap[b.Number] = b
	c.blocksList = append(c.blocksList, b)
	sort.Slice(c.blocksList, func(i, j int) bool { return c.blocksList[i].Number < c.blocksList[j].Number })
	c.mtx.Unlock()
	c.purgeBlocks(c.maxBlocksCount)
	return nil
}

func (c *DB) purgeBlocks(latestCount int64) {
	c.mtx.Lock()
	for len(c.blocksList) > int(latestCount) {
		log.Println("DELETE BLOCK", c.blocksList[0].Number)
		delete(c.blocksMap, c.blocksList[0].Number)
		c.blocksList = c.blocksList[1:]
	}
	c.mtx.Unlock()
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
