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
	//don't use parens in query when defining columns, go considers that to be
	//one value instead of 5, which throws off deserializing the result
	query := fmt.Sprintf(
		`SELECT 
		start_time, app_name, most_recent_event_time, error_count, max_idle_time
		FROM pending_sessions WHERE session_id='%s';`,
		sessionId,
	)
	row := client.Db.QueryRow(query)

	//parse the data
	var (
		startTime           uint64
		appName             string
		mostRecentEventTime uint64
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
