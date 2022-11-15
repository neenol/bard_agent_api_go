package dataService

import (
	"github.com/gin-gonic/gin"

	"launch_school/bard_agent_api/src/dataService/rabbit"
	bard "launch_school/bard_agent_api/src/structs"
)

func HandleEvents(c *gin.Context, body bard.RecordBody) error {
	if err := rabbit.SendEventsToQueue(body); err != nil {
		return err
	} else {
		return nil
	}
}
