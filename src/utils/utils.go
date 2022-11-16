package utils

import (
	"errors"
	"fmt"
	bard "launch_school/bard_agent_api/src/structs"
)

func GetTimestampFromEvent(event map[string]interface{}) (uint64, error) {
	timestampFloat, ok := event["timestamp"].(float64)
	if !ok {
		return 0, errors.New("failed to parse timestamp from event")
	}
	timestamp := uint64(timestampFloat)
	return timestamp, nil
}

func GetCountOfNewErrorsFromBody(body bard.RecordBody) uint16 {
	events := body.Events
	var newErrCount uint16 = 0
	for i := 0; i < len(events); i++ {
		event := events[i]
		if isError(event) {
			newErrCount++
		}
	}
	return newErrCount
}

/*
parse the event to see if its an error. Expecting event objects that
are errors to look like this (parts we don't are about here are omitted)

	{
		data: {
			payload: {
				level: "error"
			}
		}
		type: 6
	}

There's probably a better way to parse this, but I found this one first.
*/
func isError(event map[string]interface{}) bool {
	//use a bunch of type assertions to tell go that these keys in the
	//event map contain other maps. If the keys don't have values, then
	//the result is just nil.
	eventData := event["data"].(map[string]interface{})
	if eventData["payload"] == nil {
		return false
	}
	dataPayload := eventData["payload"].(map[string]interface{})
	if dataPayload["level"] == nil {
		return false
	}
	payloadLevel := dataPayload["level"]

	eventType := uint64(event["type"].(float64))
	fmt.Println("eventType", eventType, "payload level", payloadLevel)
	return eventType == 6 && payloadLevel == "error"
}
