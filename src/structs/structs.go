package structs

type RecordBody struct {
	SessionId     string                   `binding:"required"`
	Events        []map[string]interface{} `binding:"required"`
	MAX_IDLE_TIME uint32                   `binding:"required"`
}

type QueueMessage struct {
	SessionId string `json:"sessionId" binding:"required"`
	Event     string `json:"event" binding:"required"`
}

type PgSessionMetadata struct {
	SessionId           string
	StartTime           uint64
	AppName             string
	MostRecentEventTime uint64
	ErrorCount          uint16
	MaxIdleTime         uint32
}

type SessionMetadata struct {
	IsInPg     bool
	IsInCh     bool
	PgMetadata PgSessionMetadata
}
