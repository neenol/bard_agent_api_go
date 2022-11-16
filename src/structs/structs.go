package structs

// odd, looks like properties need to have the first letter capitalized
// in order for gin's binding functions to work...
// []map[string]interface{} means that its an array of map values. the
// maps have strings as keys. interface{} means the value can be anything
type RecordBody struct {
	SessionId     string                   `binding:"required"`
	Events        []map[string]interface{} `binding:"required"`
	MAX_IDLE_TIME uint32                   `binding:"required"`
}

// this is why go allows us to use different names for json keys and struct keys:
// keys in structs need to be capitalized.
type QueueMessage struct {
	SessionId string `json:"sessionId" binding:"required"`
	Event     string `json:"event" binding:"required"`
}

type SessionMetadata struct {
	IsInPg bool
	IsInCh bool
	//TODO: change the type when we know how the pg package works
	PgMetadata string
}
