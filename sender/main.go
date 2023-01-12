package sender

import (
	"context"
	"encoding/json"
	"log"
	"time"

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

func SendMessage(m Message) {
	ELKClient, err := elk.NewELKClient("localhost", "9200")
	if err != nil {
		log.Fatal(err)
	}

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"test-queue", // name
		false,        // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	)
	failOnError(err, "Failed to declare a queue")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	m.Timestamp = time.Now()
	jsonMessage, _ := json.Marshal(&m)

	//body := "{\"senderName\":\"Kien\",\"senderEmail\":\"641741@student.inholland.nl\",\"receiverName\":\"Kien\",\"receiverEmail\":\"kienguyen01@gmail.com\",\"body\":\"Description\",\"subject\":\"Subject\"}"

	err = ch.PublishWithContext(ctx,
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(jsonMessage),
		})

	//to elk
	elkMessage := transformELKMessage(m)
	errElk := elk.SendMessageToELK(ELKClient, &elkMessage, "message")
	if errElk != nil {
		log.Printf("Error elk: %s", errElk)
	}

	failOnError(err, "Failed to publish a message")
	log.Printf(" [x] Sent %s\n", jsonMessage)
}

func SendMulltipleMessages(m MultipleReceiverMessage) {
	ELKClient, err := elk.NewELKClient("localhost", "9200")
	if err != nil {
		log.Fatal(err)
	}

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"message", // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	failOnError(err, "Failed to declare a queue")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	m.Timestamp = time.Now()

	output := transformMessage(m)

	for _, msg := range output {
		jsonMessage, _ := json.Marshal(&msg)

		err = ch.PublishWithContext(ctx,
			"",     // exchange
			q.Name, // routing key
			false,  // mandatory
			false,  // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(jsonMessage),
			})

		log.Printf("[x] Sent  %s", msg)

		//to elk
		elkMessage := transformELKMessage(msg)
		err := elk.SendMessageToELK(ELKClient, &elkMessage, "message")
		if err != nil {
			log.Printf("Error elk: %s", err)
		}

	}

	failOnError(err, "Failed to publish a message")
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
