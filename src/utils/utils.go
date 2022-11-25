package utils

import (
	"errors"
	bard "launch_school/bard_agent_api/src/structs"
)

func GetTimestampFromEvent(event map[string]interface{}) (int64, error) {
	//go knows that event is a map with string keys that can have anything
	//as the values. but it doesn't know what keys are in it, or the types
	//of the values. This type assertion tells go to parse the value of the
	//"timestamp" key into a float64. if the "timestamp" key doesn't exist,
	//or its value can't be parsed to a float64, throw an error.
	timestampFloat, ok := event["timestamp"].(float64)
	if !ok {
		return 0, errors.New("failed to parse timestamp from event")
	}
	timestamp := int64(timestampFloat)
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
	return eventType == 6 && payloadLevel == "error"
}
