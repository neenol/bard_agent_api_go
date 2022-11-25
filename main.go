package main

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/joho/godotenv"

	"launch_school/bard_agent_api/src/bardDataService"
	auth "launch_school/bard_agent_api/src/middleware"
	bard "launch_school/bard_agent_api/src/structs"
	// TODO: restructure repo so that this module path works
	// "github.com/neenol/bard_agent_api_go/src/structs"
	// "github.com/neenol/bard_agent_api_go/src/structs"
)

func main() {
	//load env variables for use throughout the program
	if err := godotenv.Load(".env"); err != nil {
		fmt.Println("ERROR: failed to load environment variables.")
	}
	r := gin.Default()

	//initialize our database connections
	dataService, err := bardDataService.Init()
	if err != nil {
		fmt.Println("error! couldn't connect to databases")
		panic(err)
	}

	//recover from code panics by sending a 500 status request
	r.Use(gin.Recovery())

	r.GET("/authenticate", func(c *gin.Context) {
		//create a token
		token := jwt.New(jwt.SigningMethodHS256)
		claims := token.Claims.(jwt.MapClaims)
		claims["authorized"] = true
		claims["username"] = "agent"

		//sign the token
		tokenString, err := token.SignedString([]byte(os.Getenv("ACCESS_TOKEN_SECRET")))
		if err != nil {
			send500Res(c, "failed to create jwt")
			return
		}

		//send a 200 http response
		c.JSON(200, gin.H{
			"accessToken": tokenString,
		})
	})

	//use middleware to authenticate tokens before handling events
	r.POST("/record", auth.AuthenticateToken(), func(c *gin.Context) {
		//tried to use bindHeader to do this and couldn't get it to work
		appName := c.GetHeader("appname")
		if appName == "" {
			send404Res(c, "Bad request: no appname header")
			return
		}

		//get the body. BindJSON attempts to take the request body and cram
		//it into a bard.RecordBody object. Should work as long as the body has
		//the fields the object is expecting.
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

// AbortWithStatusJSON will send the response prematurely
func send404Res(c *gin.Context, msg string) {
	c.AbortWithStatusJSON(404, msg)
}

func send500Res(c *gin.Context, msg string) {
	c.AbortWithStatusJSON(500, msg)
}

func send200Res(c *gin.Context, msg string) {
	c.JSON(200, msg)
}
