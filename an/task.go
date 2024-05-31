package an

import "github.com/ipoluianov/aneth_blocks_provider/db"

type Task struct {
	Code  string
	Fn    func(result *Result, txsByMin *db.TxsByMinutes, txs []*db.Tx)
	State TaskState
}

type TaskState struct {
	Code                   string
	LastExecTime           string
	LastExecTimeDurationMs int
}

func NewTask(code string, fn func(result *Result, txsByMin *db.TxsByMinutes, txs []*db.Tx)) *Task {
	var c Task
	c.Code = code
	c.Fn = fn
	c.State.Code = code
	return &c
}
