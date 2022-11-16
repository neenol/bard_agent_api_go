package postgres

//database/sql is a go skeleton package that has unimplemented methods for talking to dbs
//lib/pq package tells database/sql the specifics of how to talk to postgres
import (
	"database/sql"
	"fmt"
	"os"

	bard "launch_school/bard_agent_api/src/structs"

	_ "github.com/lib/pq"
)

type Client struct {
	Db *sql.DB
}

func connect() (Client, error) {
	client := Client{}
	host := os.Getenv("PGHOST")
	port := os.Getenv("PGPORT")
	user := os.Getenv("PGUSER")
	password := os.Getenv("PGPASSWORD")
	dbname := os.Getenv("PGDATABASE")
	connectString := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)
	db, err := sql.Open("postgres", connectString)
	if err != nil {
		return client, err
	}
	if err := db.Ping(); err != nil {
		return client, err
	}
	fmt.Println("connected to the pg database!")
	client.Db = db
	return client, nil
}

func (c *Client) disconnect() {
	c.Db.Close()
}

func GetSessionMetadata(sessionId string) (bard.PgSessionMetadata, error) {
	metadata := bard.PgSessionMetadata{}
	client, err := connect()
	if err != nil {
		return metadata, err
	}
	defer client.disconnect()

	//TODO: sanitize db input
	query := fmt.Sprintf(
		`SELECT 
		(start_time, app_name, most_recent_event_time, error_count, max_idle_time) 
		FROM pending_sessions WHERE session_id='%s';`,
		sessionId,
	)
	fmt.Println("pg query", query)
	row := client.Db.QueryRow(query)
	fmt.Println("row from query", row)
	//parse the data
	var (
		startTime           uint32
		appName             string
		mostRecentEventTime uint32
		errorCount          uint16
		maxIdleTime         uint32
	)
	if err := row.Scan(
		&startTime,
		&appName,
		&mostRecentEventTime,
		&errorCount,
		&maxIdleTime,
	); err != nil {
		fmt.Println("didn't find a row in postgres!")
		//don't actually throw an error if we didn't find a session
		return metadata, nil
	}
	metadata.SessionId = sessionId
	metadata.StartTime = startTime
	metadata.AppName = appName
	metadata.MostRecentEventTime = mostRecentEventTime
	metadata.ErrorCount = errorCount
	metadata.MaxIdleTime = maxIdleTime
	return metadata, nil
}
