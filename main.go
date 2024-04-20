package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

func main() {
	conn, err := connect()
	if err != nil {
		panic((err))
	}

	ctx := context.Background()
	rows, err := conn.Query(ctx, "SELECT id FROM t1 LIMIT 5")
	if err != nil {
		log.Fatal(err)
	}

	for rows.Next() {
		var (
			id int32
		)
		if err := rows.Scan(
			&id,
		); err != nil {
			log.Fatal(err)
		}
		fmt.Println("ROW: ", id)
	}

}

func connect() (driver.Conn, error) {
	var (
		ctx       = context.Background()
		conn, err = clickhouse.Open(&clickhouse.Options{
			Addr: []string{"localhost:9000"},
			Auth: clickhouse.Auth{
				Database: "default",
				Username: "default",
				Password: "bbgj9",
			},
			ClientInfo: clickhouse.ClientInfo{
				Products: []struct {
					Name    string
					Version string
				}{
					{Name: "an-example-go-client", Version: "0.1"},
				},
			},

			Debugf: func(format string, v ...interface{}) {
				fmt.Printf(format, v)
			},
		})
	)

	if err != nil {
		return nil, err
	}

	if err := conn.Ping(ctx); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			fmt.Printf("Exception [%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace)
		}
		return nil, err
	}
	return conn, nil
}
