package clickhouse

import (
	"context"
	"fmt"
	"os"

	ch "github.com/ClickHouse/clickhouse-go/v2"
)

type Client struct {
	Conn ch.Conn
}

func DoesSessionExist(sessionId string) (bool, error) {
	client, err := connect()
	if err != nil {
		return false, err
	}
	defer client.disconnect() //disconnect when we're done with this fn

	//TODO: sanitize the query
	row := client.Conn.QueryRow(
		context.Background(),
		fmt.Sprintf("SELECT sessionId FROM sessionTable WHERE sessionId='%s'", sessionId),
	)

	var col1 string
	//Scan returns an error if we don't find the row
	if err := row.Scan(&col1); err != nil {
		return false, nil
	} else {
		return true, nil
	}
}

func connect() (Client, error) {
	conn, err := ch.Open(&ch.Options{
		//go client uses TCP instead of http, so use port 9000 instead of 8123
		Addr: []string{fmt.Sprintf("%s:%d", os.Getenv("CLICKHOUSE_HOST"), 9000)},
		Auth: ch.Auth{
			Database: "eventDb",
			Username: "default",
			Password: "",
		},
	})
	if err != nil {
		return Client{}, err
	}
	client := Client{Conn: conn}
	fmt.Println("connected to clickhouse!")
	return client, nil
}

func (c *Client) disconnect() {
	c.Conn.Close()
	fmt.Println("disconnected from clickhouse")
}
