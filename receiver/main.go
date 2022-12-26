package main

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

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
		"",          // consumer
		true,        // auto-ack
		false,       // exclusive
		false,       // no-local
		false,       // no-wait
		nil,         // args
	)

	//process messages from this channel
	//forever will BLOCK the main.go program until it received a value from the channel
	forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}