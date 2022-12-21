package bardDataService

import (
	"github.com/gin-gonic/gin"

	"launch_school/bard_agent_api/src/bardDataService/postgres"
	"launch_school/bard_agent_api/src/bardDataService/rabbit"
	"launch_school/bard_agent_api/src/bardDataService/redis"
	bard "launch_school/bard_agent_api/src/structs"
	"launch_school/bard_agent_api/src/utils"
)

// define a local DataService struct so I can attach methods to it here.
type DataService struct {
	Postgres postgres.Client
	Rabbit   rabbit.Client
	Redis    redis.Client
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
	redisClient := redis.ConnectToCaches()
	dataService.Postgres = postgresClient
	dataService.Rabbit = rabbitClient
	dataService.Redis = redisClient
	return dataService, nil
}

func (ds DataService) HandleEvents(c *gin.Context, body bard.RecordBody, appName string) error {
	sessionId := body.SessionId
	if ds.Redis.IsActiveSession(sessionId) {
		//existing session
		return ds.updateExistingSession(body)
	} else if ds.Redis.IsEndedSession(sessionId) {
		//ended session
		return nil
	} else {
		//new session
		return ds.Postgres.CreateNewSession(body, appName, ds.updateExistingSession)
	}
}

/*	PRIVATE METHODS	*/
func (ds DataService) updateExistingSession(body bard.RecordBody) error {
	//update most recent event time in the cache
	mostRecentEventTime, err := utils.GetTimestampFromEvent(body.Events[0])
	if err != nil {
		return err
	}
	sessionId := body.SessionId
	if err := ds.Redis.UpdateMostRecentEventTime(sessionId, mostRecentEventTime); err != nil {
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
