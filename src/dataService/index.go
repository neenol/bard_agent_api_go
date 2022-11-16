package dataService

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"launch_school/bard_agent_api/src/dataService/clickhouse"
	"launch_school/bard_agent_api/src/dataService/postgres"
	"launch_school/bard_agent_api/src/dataService/rabbit"
	bard "launch_school/bard_agent_api/src/structs"
)

// test for this:
// 4071c287-4b45-4367-8591-go-session  is in ch and pg. (ended)
// 4071c287-4b45-4367-8591-node-session is just in ch. (ended)
// 4071c287-4b45-4367-8590- is just in pg (existing)
// bluh is in postgres
func HandleEvents(c *gin.Context, body bard.RecordBody, appName string) error {
	metadata, err := getSessionMetadata(body.SessionId)
	if err != nil {
		return err
	}
	if isNewSession(metadata) {
		fmt.Println("new session")
		return postgres.CreateNewSession(body, appName)
	} else if isEndedSession(metadata) {
		fmt.Println("ended session")
	} else {
		fmt.Println("existing session")
		//update most recent event time
		//update eror count
		if err := rabbit.SendEventsToQueue(body); err != nil {
			return err
		}
	}

	return nil
}

func getSessionMetadata(sessionId string) (bard.SessionMetadata, error) {
	var metadata = bard.SessionMetadata{}
	//TODO: change this
	//postgres metadata
	pgMetadata, err := postgres.GetSessionMetadata(sessionId)
	if err != nil {
		return metadata, err
	}
	var isInPg bool
	if pgMetadata.SessionId != "" {
		isInPg = true
	}

	//clickhouse metadata
	isInCh, err := clickhouse.DoesSessionExist(sessionId)
	if err != nil {
		return metadata, err
	}
	metadata.IsInPg = isInPg
	metadata.IsInCh = isInCh
	metadata.PgMetadata = pgMetadata
	return metadata, nil
}

func isNewSession(metadata bard.SessionMetadata) bool {
	return metadata.IsInPg == false && metadata.IsInCh == false
}

func isEndedSession(metadata bard.SessionMetadata) bool {
	return metadata.IsInCh == true
}
