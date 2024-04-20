package db

import (
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
)

func CreateDB() {
	conn := clickhouse.OpenDB(&clickhouse.Options{
		Addr: []string{"127.0.0.1:8123"},
		Auth: clickhouse.Auth{
			Database: "default",
			Username: "default",
			Password: "",
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		DialTimeout: 30 * time.Second,
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
		Protocol: clickhouse.HTTP,
	})
	fmt.Println(conn, 0)
}
