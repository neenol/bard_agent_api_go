package rabbit

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	bard "launch_school/bard_agent_api/src/structs"
)

// attach rabbit methods to another client struct. makes for neater
// syntax when I use those methods in the index.go data service file.
type Client struct {
	Connection *amqp.Connection
	Channel    *amqp.Channel
}

func InitClient() (Client, error) {
	client := Client{}
	rabbit, err := amqp.Dial(fmt.Sprintf("amqp://%s:5672", os.Getenv("RABBITMQ_HOST")))
	if err != nil {
		return client, err
	}
	channel, err := rabbit.Channel()
	if err != nil {
		return client, err
	}
	if err := channel.ExchangeDeclare(
		"test-exchange",
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return client, err
	}
	client.Channel = channel
	client.Connection = rabbit
	fmt.Println("Rabbit exchange created!")
	return client, nil
}

func (client Client) SendEventsToQueue(body bard.RecordBody) error {

	sessionId := body.SessionId
	events := body.Events
	for i := 0; i < len(events); i++ {
		//The queue needs a byte array that represents a stringified json object with the following format
		// {
		// 	sessionId: string,
		// 	event: string
		// }
		//the event is a stringified json representation of the event.

		//first, build the stringified json. marshall creates a byte array of the json string.
		var eventMap = events[i]
		encodedEventJsonString, err := json.Marshal(eventMap)
		if err != nil {
			return err
		}
		//turn the byte array into the json string it represents
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
		if err := client.Channel.PublishWithContext(
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
	}
	return nil
}
