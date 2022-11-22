package bardDataService

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"launch_school/bard_agent_api/src/bardDataService/postgres"
	"launch_school/bard_agent_api/src/bardDataService/rabbit"
	"launch_school/bard_agent_api/src/bardDataService/redis"
	bard "launch_school/bard_agent_api/src/structs"
	"launch_school/bard_agent_api/src/utils"
)

// test for this:
// 4071c287-4b45-4367-8591-go-session  is in ch and pg. (ended)
// 4071c287-4b45-4367-8591-node-session is just in ch. (ended)
// 4071c287-4b45-4367-8590- is just in pg (existing)
// bluh is in postgres
type DataService struct {
	Postgres postgres.Client
	Rabbit   rabbit.Client
}

/*	PUBLIC METHODS	*/
func Init() (DataService, error) {
	dataService := DataService{}
	postgresClient, err := postgres.Connect()
	if err != nil {
		return dataService, err
	}
	rabbitClient, err := rabbit.InitClient()
	if err != nil {
		return dataService, err
	}
	dataService.Postgres = postgresClient
	dataService.Rabbit = rabbitClient
	return dataService, nil
}
func (ds DataService) HandleEvents(c *gin.Context, body bard.RecordBody, appName string) error {
	sessionId := body.SessionId
	if redis.IsActiveSession(sessionId) {
		fmt.Println("existing session")
		return ds.updateExistingSession(body)
	} else if redis.IsEndedSession(sessionId) {
		fmt.Println("ended session")
		return nil
	} else {
		fmt.Println("new session")
		return ds.Postgres.CreateNewSession(body, appName, ds.updateExistingSession)
	}
}

func (ds DataService) updateExistingSession(body bard.RecordBody) error {
	//update most recent event time
	if err := ds.Postgres.UpdateMostRecentEventTime(body); err != nil {
		return err
	}
	//update error count
	if err := ds.updateErrorCount(body); err != nil {
		return err
	}
	if err := ds.Rabbit.SendEventsToQueue(body); err != nil {
		return err
	}

	return nil
}

func (ds DataService) updateErrorCount(body bard.RecordBody) error {
	//parse the new number of errors from the events
	newErrorCount := utils.GetCountOfNewErrorsFromBody(body)
	if newErrorCount == 0 {
		return nil
	}

	//increment count of errors in postgres if there're new ones
	sessionId := body.SessionId
	return ds.Postgres.IncrementErrorCount(sessionId, newErrorCount)
}
