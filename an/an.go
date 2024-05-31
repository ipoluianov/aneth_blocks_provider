package an

import (
	"sync"
	"time"

	"github.com/ipoluianov/aneth_blocks_provider/db"
	"github.com/ipoluianov/gomisc/logger"
)

type An struct {
	analytics map[string]*Result
	mtx       sync.Mutex
	tasks     []*Task
}

type AnState struct {
	Tasks []*TaskState
}

var Instance *An

func init() {
	Instance = NewAn()
}

func NewAn() *An {
	var c An
	c.analytics = make(map[string]*Result)
	c.tasks = append(c.tasks, NewTask("minutes_count", c.taskMinutesCount))
	c.tasks = append(c.tasks, NewTask("minutes_values", c.taskMinutesValues))
	return &c
}

func (c *An) Start() {
	go c.ThAn()
}

func (c *An) GetState() *AnState {
	var st AnState
	c.mtx.Lock()
	for _, t := range c.tasks {
		state := t.State
		st.Tasks = append(st.Tasks, &state)
	}
	c.mtx.Unlock()
	return &st
}

func (c *An) GetResult(code string) *Result {
	var res *Result
	c.mtx.Lock()
	defer c.mtx.Unlock()
	if r, ok := c.analytics[code]; ok {
		res = r
	}
	return res
}

func (c *An) ThAn() {
	for {
		dt1 := time.Now()
		logger.Println("")
		logger.Println("---------- an ------------")
		logger.Println("reading transactions")
		txsByMinutes, txs := db.Instance.GetLatestTransactions()
		count := 0
		for _, item := range txsByMinutes.Items {
			count += len(item.TXS)
		}
		logger.Println("An::an txs:", len(txs))

		for _, task := range c.tasks {
			var res Result
			res.Code = task.Code
			dtBegin := time.Now()
			task.Fn(&res, txsByMinutes, txs)
			dtEnd := time.Now()
			duration := dtEnd.Sub(dtBegin).Milliseconds()
			task.State.Code = task.Code
			task.State.LastExecTime = time.Now().Format("2006-01-02 15:04:05")
			task.State.LastExecTimeDurationMs = int(duration)
			c.mtx.Lock()
			c.analytics[task.Code] = &res
			c.mtx.Unlock()
		}

		dt2 := time.Now()
		logger.Println("execution time:", dt2.Sub(dt1).Milliseconds(), "ms")
		logger.Println("--------------------------")
		logger.Println("")

		time.Sleep(3 * time.Second)
	}
}
