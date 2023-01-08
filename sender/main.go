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
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func SendMessage(m Message) {
	client, err := elk.NewELKClient("localhost", "9200")
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

	client.SendLog("sending", jsonMessage)
	failOnError(err, "Failed to publish a message")
	log.Printf(" [x] Sent %s\n", jsonMessage)
}
