package redis

import (
	"context"
	"fmt"

	r "github.com/go-redis/redis/v8"
)

const ACTIVE_SESSION_DB = 1
const ENDED_SESSION_DB = 2

/*	PUBLIC METHODS	*/
func IsActiveSession(sessionId string) bool {
	return isKeyInRedis(ACTIVE_SESSION_DB, sessionId)
}

func IsEndedSession(sessionId string) bool {
	return isKeyInRedis(ENDED_SESSION_DB, sessionId)
}

func AddSessionToActiveCache(sessionId string) error {
	return addSessionToCache(ACTIVE_SESSION_DB, sessionId)
}

/*	PRIVATE METHODS	*/
func isKeyInRedis(db int, sessionId string) bool {
	client := connect(db)
	//todo: doc says that this isn't really the best way of doing things
	defer disconnect(client)
	_, err := client.Get(context.Background(), sessionId).Result()

	//if err is this special value, it means the key isn't in redis
	if err == r.Nil {
		fmt.Printf("%s isn't in redis\n", sessionId)
		return false
	} else if err != nil {
		fmt.Println("redis fetch failed!")
		panic(err)
	} else {
		return true
	}
}

func addSessionToCache(db int, sessionId string) error {
	client := connect(db)
	//todo: doc says that this isn't really the best way of doing things
	defer disconnect(client)

	_, err := client.Set(context.Background(), sessionId, true, 0).Result()
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

func disconnect(client *r.Client) {
	client.Close()
}
