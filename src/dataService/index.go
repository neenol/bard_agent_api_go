package dataService

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"launch_school/bard_agent_api/src/dataService/clickhouse"
	"launch_school/bard_agent_api/src/dataService/rabbit"
	bard "launch_school/bard_agent_api/src/structs"
)

func HandleEvents(c *gin.Context, body bard.RecordBody) error {
	metadata, err := getSessionMetadata(body.SessionId)
	if err != nil {
		return err
	}
	fmt.Println("metadata", metadata)

	if err := rabbit.SendEventsToQueue(body); err != nil {
		return err
	}

	return nil
}

func getSessionMetadata(sessionId string) (bard.SessionMetadata, error) {
	var metadata = bard.SessionMetadata{}
	//TODO: change this
	//postgres metadata
	pgMetadata := "placeholder metadata"
	var isInPg bool
	if pgMetadata != "" {
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
