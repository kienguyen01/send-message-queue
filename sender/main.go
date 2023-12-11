package sender

import (
	"context"
	"encoding/json"
	"log"
	"time"
	"os"

	elk "github.com/kienguyen01/send-message-queue/elk"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Message struct {
	SenderEmail   string
	SenderName    string
	ReceiverEmail string
	ReceiverName  string
	Body          string
	Subject       string
	Timestamp     time.Time
}

type MultipleReceiverMessage struct {
	SenderEmail   string
	SenderName    string
	ReceiverEmail []string
	ReceiverName  []string
	Body          string
	Subject       string
	Timestamp     time.Time
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func setupMessageQueue() (*amqp.Connection, *amqp.Channel, amqp.Queue, error) {
	conn, err := amqp.Dial(os.Getenv("RABBITMQ_HOST"))
	if err != nil {
		return nil, nil, amqp.Queue{}, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, nil, amqp.Queue{}, err
	}

	q, err := ch.QueueDeclare(
		"message", // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, nil, amqp.Queue{}, err
	}

	return conn, ch, q, nil
}

func publishMessage(ch *amqp.Channel, q amqp.Queue, message interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	jsonMessage, err := json.Marshal(message)
	if err != nil {
		return err
	}

	err = ch.PublishWithContext(ctx,
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        jsonMessage,
		})
	if err != nil {
		return err
	}

	log.Printf(" [x] Sent %s\n", jsonMessage)
	return nil
}

func SendMessage(m Message) error {
	ELKClient, err := elk.NewELKClient(os.Getenv("ELASTIC_HOST"), os.Getenv("ELASTIC_PORT"))
	if err != nil {
		log.Fatal(err)
	}

	conn, ch, q, err := setupMessageQueue()
	if err != nil {
		failOnError(err, "Failed to set up the message queue")
	}
	defer conn.Close()
	defer ch.Close()

	m.Timestamp = time.Now()
	err = publishMessage(ch, q, m)
	failOnError(err, "Failed to publish a message")

	// to ELK
	elkMessage := transformELKMessage(m)
	err = elk.SendMessageToELK(ELKClient, &elkMessage, "message")
	if err != nil {
		log.Printf("Error elk: %s", err)
	}

	return err
}

func SendMultipleMessages(m MultipleReceiverMessage) error {
	ELKClient, err := elk.NewELKClient(os.Getenv("ELASTIC_HOST"), os.Getenv("ELASTIC_PORT"))
	if err != nil {
		log.Fatal(err)
	}

	conn, ch, q, err := setupMessageQueue()
	if err != nil {
		failOnError(err, "Failed to set up the message queue")
	}
	defer conn.Close()
	defer ch.Close()

	m.Timestamp = time.Now()
	output := transformMessage(m)

	for _, msg := range output {
		err = publishMessage(ch, q, msg)
		if err != nil {
			failOnError(err, "Failed to publish a message")
		}

		// to ELK
		elkMessage := transformELKMessage(msg)
		err = elk.SendMessageToELK(ELKClient, &elkMessage, "message")
		if err != nil {
			log.Printf("Error elk: %s", err)
		}
	}

	return err
}

func transformMessage(msg MultipleReceiverMessage) []Message {
	// Create a slice to hold the messages
	messages := []Message{}

	// Loop through the receivers
	for i := 0; i < len(msg.ReceiverEmail); i++ {
		// Create a new Message
		newMsg := Message{
			SenderEmail:   msg.SenderEmail,
			SenderName:    msg.SenderName,
			ReceiverEmail: msg.ReceiverEmail[i],
			ReceiverName:  msg.ReceiverName[i],
			Body:          msg.Body,
			Subject:       msg.Subject,
			Timestamp:     msg.Timestamp,
		}

		// Append the new message to the slice
		messages = append(messages, newMsg)
	}

	return messages
}

func transformELKMessage(msg Message) elk.Message {

	newMsg := elk.Message{
		SenderEmail:   msg.SenderEmail,
		SenderName:    msg.SenderName,
		ReceiverEmail: msg.ReceiverEmail,
		ReceiverName:  msg.ReceiverName,
		Body:          msg.Body,
		Subject:       msg.Subject,
		Timestamp:     msg.Timestamp,
	}

	return newMsg
}
