package postgres

import (
	"database/sql"
	"fmt"
	"os"

	bard "launch_school/bard_agent_api/src/structs"
	"launch_school/bard_agent_api/src/utils"

	_ "github.com/lib/pq"
)

// attach postgres methods to another client struct. makes for neater
// syntax when I use those methods in the index.go data service file.
type Client struct {
	Db *sql.DB
}

/*	PUBLIC METHODS	*/
func Connect() (Client, error) {
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
	client.Db = db
	fmt.Println("Connected to postgres!")
	return client, nil
}

func (client Client) CreateNewSession(
	body bard.RecordBody,
	appName string,
	updateExistingSession func(bard.RecordBody) error,
) error {

	sessionId := body.SessionId
	startTime, err := utils.GetTimestampFromEvent(body.Events[0])
	if err != nil {
		return err
	}
	MAX_IDLE_TIME := body.MAX_IDLE_TIME
	mostRecentEventTime := startTime

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

	//call this to update error counts and such after creation
	return updateExistingSession(body)
}

func (client Client) IncrementErrorCount(sessionId string, newErrorCount uint16) error {
	query := `UPDATE pending_sessions
						SET error_count = error_count + $1
						WHERE session_id = $2;
						`
	_, err := client.Db.Exec(query, newErrorCount, sessionId)
	if err != nil {
		return err
	}
	return err
}
