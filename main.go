package main

//getting an error here, but it must be something with the extension
//because the code runs just fine
import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
)

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

	r.POST("/record", func(c *gin.Context) {
		//tried to use bindHeader to do this and couldn't get it to work
		appName := c.GetHeader("appname")
		if appName == "" {
			send404Res(c, "Bad request: no appname header")
			return
		}

		//get the body
		var body RecordBody
		if err := c.BindJSON(&body); err != nil {
			msg := fmt.Sprintf("Bad request: invalid body. %s", err)
			send404Res(c, msg)
			return
		}

		//handle our events
		if err := handleEvents(c, body); err != nil {
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

func handleEvents(c *gin.Context, body RecordBody) error {
	// conn, err := clickhouse.Open(&clickhouse.Options{
	// 	Addr: []string{fmt.Sprintf("%s:%d", os.Getenv("CLICKHOUSE_HOST"), 8123)},
	// 	Auth: clickhouse.Auth{
	// 		Database: "eventDb",
	// 		Username: "default",
	// 		Password: "",
	// 	},
	// })
	// if err != nil {
	// 	return err
	// }
	fmt.Println("connected to clickhouse!")
	rabbit, err := amqp.Dial(fmt.Sprintf("amqp://%s:5672", os.Getenv("RABBITMQ_HOST")))
	if err != nil {
		return err
	}
	defer rabbit.Close()
	fmt.Println("connected to rabbit!")
	channel, err := rabbit.Channel()
	if err != nil {
		return err
	}
	fmt.Println("rabbit channel established!")
	defer channel.Close()
	if err := channel.ExchangeDeclare(
		"test-exchange",
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return err
	}
	fmt.Println("Rabbit exchange created!")
	//prepare a clickhouse batch
	// batch, err := conn.PrepareBatch(context.Background(), "INSERT INTO eventTable")
	// if err != nil {
	// 	return err
	// }
	// fmt.Println("initialized batch!")
	sessionId := body.SessionId
	events := body.Events
	for i := 0; i < len(events); i++ {
		//The queue needs a byte array that represents a stringified json object with the following format
		// {
		// 	sessionId: string,
		// 	event: string
		// }
		//the event is a stringified json representation of the event. so we need to
		// transform the event from a map into a json string
		//build a message object
		//turn that message object into a jso nstring
		//turn that string into a byte array

		//first, build the stringified json. marshall creates a byte array of the json string.
		var eventMap = events[i]
		encodedEventJsonString, err := json.Marshal(eventMap)
		if err != nil {
			return err
		}
		//turn the byte array back into a string. we need that for the message obj, NOT a byte arr
		eventJsonString := string(encodedEventJsonString)

		//build and serialize the queue message
		queueMessage := QueueMessage{sessionId, eventJsonString}
		encodedQueueMessage, err := json.Marshal(queueMessage)
		if err != nil {
			return err
		}

		//publish the event
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := channel.PublishWithContext(
			ctx,
			"test-exchange",
			"",
			false,
			false,
			amqp.Publishing{
				ContentType:  "application/json",
				DeliveryMode: amqp.Persistent,
				Body:         encodedQueueMessage,
			},
		); err != nil {
			return err
		}
		//build a batch to send all event data to ch at once.
		// if err := batch.Append(
		// 	sessionId,
		// 	event,
		// ); err != nil {
		// 	return err
		// }
	}

	// //send off our built batch
	// if err := batch.Send(); err != nil {
	// 	return err
	// }
	return nil
}
