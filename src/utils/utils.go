package utils

import (
	"errors"
)

func GetTimestampFromEvent(event map[string]interface{}) (uint64, error) {
	timestampFloat, ok := event["timestamp"].(float64)
	if !ok {
		return 0, errors.New("failed to parse timestamp from event")
	}
	timestamp := uint64(timestampFloat)
	return timestamp, nil
}
