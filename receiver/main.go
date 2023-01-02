package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type Message struct {
	SenderEmail   string
	SenderName    string
	ReceiverEmail string
	ReceiverName  string
	Body          string
	Subject       string
}

func main() {
	log.Println("Consumer - Connecting to the Rabbit channel")

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")

	if err != nil {
		log.Println("Error in connection")
	}

	//close connection at end of program
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}

	defer ch.Close()

	msgs, err := ch.Consume(
		"test-queue", // queue
		"",           // consumer
		true,         // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)

	//process messages from this channel
	//forever will BLOCK the main.go program until it received a value from the channel
	forever := make(chan bool)

	go func() {
		for d := range msgs {
			var m Message
			json.Unmarshal(d.Body, &m)

			SendEmail(m)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

func SendEmail(m Message) {
	from := mail.NewEmail("Kien Nguyen", "641741@student.inholland.nl")
	subject := "Sending with SendGrid is Fun"
	to := mail.NewEmail("Example User", "kienguyen01@gmail.com")
	plainTextContent := "and easy to do anywhere, even with Go " + m.Body
	htmlContent := "<strong>and easy to do anywhere, even with Go</strong>"
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
	response, err := client.Send(message)
	if err != nil {
		log.Println(err)
	} else {
		fmt.Println(response.StatusCode)
		fmt.Println(response.Body)
		fmt.Println(response.Headers)
	}
}
