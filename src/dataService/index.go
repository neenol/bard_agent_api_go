package dataService

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"

	bard "launch_school/bard_agent_api/src/structs"
)

func HandleEvents(c *gin.Context, body bard.RecordBody) error {
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
		queueMessage := bard.QueueMessage{SessionId: sessionId, Event: eventJsonString}
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
