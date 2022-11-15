package main

//getting an error here, but it must be something with the extension
//because the code runs just fine
import (
	"fmt"

	"github.com/gin-gonic/gin"
)

//TODO: better typing for events
//odd, looks like properties need to have the first letter capitalized
//in order for gin's binding functions to work...
type RecordBody struct {
	SessionId string `binding:"required"`;
	Events []string ` binding:"required"`;
	MAX_IDLE_TIME uint32 `binding:"required"`;
}

func main() {
	r := gin.Default()
	//basic path
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	//getting a url path param
	r.GET("/ping/:payload", func(c *gin.Context) {
		payload:= c.Param("payload")
		c.JSON(200, gin.H{
			"payload": payload,
		})
		fmt.Println("payload is ", payload)
	})
	//getting string query params: expecting /query?database=postgres&queue=rabbit
	r.GET("/query", func(c *gin.Context) {
		//gets the query value for 'database' and returns the 2nd arg if its not there
		db := c.DefaultQuery("database", "N/A")
		queue := c.Query("queue")
		fmt.Println("db is", db, "and queue is", queue)
	})

	r.POST("/record", func(c *gin.Context) {
		//tried to use bindHeader to do this and couldn't get it to work
		appName := c.GetHeader("appname")
		if appName == "" {
			send404Res(c, "bad request: no appname header")
			return
		}

		//get the body
		var body RecordBody
		if err := c.BindJSON(&body); err != nil {
			msg :=fmt.Sprintf("Bad request: invalid body. %s", err)
			send404Res(c, msg)
			return
		}
		sessionId := body.SessionId
		events := body.Events
		maxIdleTime := body.MAX_IDLE_TIME
		fmt.Println("the gangs all here!", sessionId, events, maxIdleTime)
	})
	r.Run(":3001")
}

func send404Res(c *gin.Context, msg string) {
	c.JSON(404, msg)
}