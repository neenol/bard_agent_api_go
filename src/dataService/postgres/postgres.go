package postgres

//database/sql is a go skeleton package that has unimplemented methods for talking to dbs
//lib/pq package tells database/sql the specifics of how to talk to postgres
import (
	"database/sql"
	"fmt"
	"os"
	"time"

	bard "launch_school/bard_agent_api/src/structs"
	"launch_school/bard_agent_api/src/utils"

	_ "github.com/lib/pq"
)

type Client struct {
	Db *sql.DB
}

/*	PUBLIC METHODS	*/
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
	query := `SELECT 
		start_time, app_name, most_recent_event_time, error_count, max_idle_time
		FROM pending_sessions WHERE session_id=$1;`
	row := client.Db.QueryRow(query, sessionId)

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

func CreateNewSession(body bard.RecordBody, appName string) error {
	client, err := connect()
	if err != nil {
		return err
	}
	defer client.disconnect()

	sessionId := body.SessionId
	startTime, err := utils.GetTimestampFromRecordBody(body)
	if err != nil {
		return err
	}
	MAX_IDLE_TIME := body.MAX_IDLE_TIME
	mostRecentEventTime := time.Now().UnixMilli()

	query := `INSERT INTO pending_sessions
						(session_id, start_time, most_recent_event_time, app_name, max_idle_time)
						VALUES
						($1, $2, $3, $4, $5);`
	_, err = client.Db.Exec(
		query,
		sessionId,
		startTime,
		mostRecentEventTime,
		appName,
		MAX_IDLE_TIME,
	)
	if err != nil {
		return err
	}
	return nil
}

/*	PRIVATE METHODS	*/
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
