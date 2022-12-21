package redis

import (
	"context"
	"fmt"

	r "github.com/go-redis/redis/v8"
)

// we've got TWO redis databases in play here: one for active sessions, one
// for inactive sessions
const ACTIVE_SESSION_DB = 1
const ENDED_SESSION_DB = 2

type Client struct {
	ActiveCache *r.Client
	EndedCache  *r.Client
}

/*	PUBLIC METHODS	*/
func ConnectToCaches() Client {
	activeCache := connect(ACTIVE_SESSION_DB)
	endedCache := connect(ENDED_SESSION_DB)
	client := Client{}
	client.ActiveCache = activeCache
	client.EndedCache = endedCache
	fmt.Println("Connected to Redis!")
	return client
}

func (client Client) IsActiveSession(sessionId string) bool {
	return client.isKeyInRedis(sessionId, true)
}
func (client Client) IsEndedSession(sessionId string) bool {
	return client.isKeyInRedis(sessionId, false)
}

func (client Client) UpdateMostRecentEventTime(
	sessionId string,
	mostRecentEventTime int64,
) error {
	return client.updateActiveSession(sessionId, mostRecentEventTime)
}

/*	PRIVATE METHODS	*/
func (client Client) isKeyInRedis(
	sessionId string,
	checkingActiveCache bool,
) bool {
	cache := client.getCache(checkingActiveCache)

	_, err := cache.Get(context.Background(), sessionId).Result()

	//if err is this special value, it means the key isn't in redis
	if err == r.Nil {
		return false
	} else if err != nil {
		fmt.Println("redis fetch failed!")
		panic(err)
	} else {
		return true
	}
}

func (client Client) getCache(checkingActiveCache bool) *r.Client {
	if checkingActiveCache {
		return client.ActiveCache
	} else {
		return client.EndedCache
	}
}

func (client Client) updateActiveSession(
	sessionId string,
	mostRecentEventTimestamp int64,
) error {
	//don't update time for ended sessions
	if client.IsEndedSession(sessionId) {
		return nil
	}

	//if the session is active, update it
	_, err := client.ActiveCache.Set(
		context.Background(),
		sessionId,
		mostRecentEventTimestamp,
		0,
	).Result()
	return err
}

func connect(db int) *r.Client {
	client := r.NewClient(&r.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       db,
	})
	return client
}
