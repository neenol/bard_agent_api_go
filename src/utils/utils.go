package utils

import (
	"errors"
	bard "launch_school/bard_agent_api/src/structs"
)

func GetTimestampFromRecordBody(body bard.RecordBody) (uint64, error) {
	event := body.Events[0]
	timestampFloat, ok := event["timestamp"].(float64)
	if !ok {
		return 0, errors.New("failed to parse timestamp from event")
	}
	timestamp := uint64(timestampFloat)
	return timestamp, nil
}
