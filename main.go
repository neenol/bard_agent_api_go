package main

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/joho/godotenv"

	"launch_school/bard_agent_api/src/dataService"
	bard "launch_school/bard_agent_api/src/structs"
	// TODO: restructure repo so that this module path works
	// "github.com/neenol/bard_agent_api_go/src/structs"
	// "github.com/neenol/bard_agent_api_go/src/structs"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		fmt.Println("ERROR: failed to load environment variables.")
	}
	r := gin.Default()
	//basic path
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	//getting a url path param
	r.GET("/ping/:payload", func(c *gin.Context) {
		payload := c.Param("payload")
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

	r.GET("/authenticate", func(c *gin.Context) {
		// user := bard.User{}
		// user.Name = "agent"
		token := jwt.New(jwt.SigningMethodHS256)
		claims := token.Claims.(jwt.MapClaims)
		claims["authorized"] = true
		claims["username"] = "agent"
		tokenString, err := token.SignedString([]byte(os.Getenv("ACCESS_TOKEN_SECRET")))
		if err != nil {
			send500Res(c, "failed to create jwt")
			return
		}
		c.JSON(200, gin.H{
			"accessToken": tokenString,
		})
	})

	r.POST("/record", func(c *gin.Context) {
		//tried to use bindHeader to do this and couldn't get it to work
		appName := c.GetHeader("appname")
		if appName == "" {
			send404Res(c, "Bad request: no appname header")
			return
		}

		//get the body
		var body bard.RecordBody
		if err := c.BindJSON(&body); err != nil {
			msg := fmt.Sprintf("Bad request: invalid body. %s", err)
			send404Res(c, msg)
			return
		}

		//handle our events
		if err := dataService.HandleEvents(c, body, appName); err != nil {
			msg := fmt.Sprintf("Event handling error. %s", err)
			send500Res(c, msg)
			return
		} else {
			send200Res(c, "thanks")
			return
		}
	})
	r.Run(":3001")
}

func send404Res(c *gin.Context, msg string) {
	c.JSON(404, msg)
}

func send500Res(c *gin.Context, msg string) {
	c.JSON(500, msg)
}

func send200Res(c *gin.Context, msg string) {
	c.JSON(200, msg)
}

func send401Res(c *gin.Context, msg string) {
	c.JSON(401, msg)
}
