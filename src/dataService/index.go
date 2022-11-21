package dataService

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"launch_school/bard_agent_api/src/dataService/postgres"
	"launch_school/bard_agent_api/src/dataService/rabbit"
	"launch_school/bard_agent_api/src/dataService/redis"
	bard "launch_school/bard_agent_api/src/structs"
	"launch_school/bard_agent_api/src/utils"
)

// test for this:
// 4071c287-4b45-4367-8591-go-session  is in ch and pg. (ended)
// 4071c287-4b45-4367-8591-node-session is just in ch. (ended)
// 4071c287-4b45-4367-8590- is just in pg (existing)
// bluh is in postgres

/*	PUBLIC METHODS	*/
func HandleEvents(c *gin.Context, body bard.RecordBody, appName string) error {
	sessionId := body.SessionId
	if redis.IsActiveSession(sessionId) {
		fmt.Println("existing session")
		return updateExistingSession(body)
	} else if redis.IsEndedSession(sessionId) {
		fmt.Println("ended session")
		return nil
	} else {
		fmt.Println("new session")
		return postgres.CreateNewSession(body, appName, updateExistingSession)
	}
}

func updateExistingSession(body bard.RecordBody) error {
	//update most recent event time
	if err := postgres.UpdateMostRecentEventTime(body); err != nil {
		return err
	}
	//update error count
	if err := updateErrorCount(body); err != nil {
		return err
	}
	if err := rabbit.SendEventsToQueue(body); err != nil {
		return err
	}

	return nil
}

func updateErrorCount(body bard.RecordBody) error {
	//parse the new number of errors from the events
	newErrorCount := utils.GetCountOfNewErrorsFromBody(body)
	sessionId := body.SessionId

	//increment count of errors in postgres
	return postgres.IncrementErrorCount(sessionId, newErrorCount)
}
