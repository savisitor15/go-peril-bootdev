package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	amqp "github.com/rabbitmq/amqp091-go"
	gamelogic "github.com/savisitor15/go-peril-bootdev/internal/gamelogic"
	"github.com/savisitor15/go-peril-bootdev/internal/pubsub"
	"github.com/savisitor15/go-peril-bootdev/internal/routing"
)

func main() {
	fmt.Println("Starting Peril client...")
	rabbitHost := "amqp://guest:guest@localhost:5672/"
	rabbitConn, err := amqp.Dial(rabbitHost)
	if err != nil {
		log.Fatalf("error creating connetion: %v", err)
	}
	defer rabbitConn.Close()
	fmt.Printf("connection esstablished to %s \n", rabbitHost)
	// prompt for username
	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("error getting username: %v", err)
	}
	fmt.Println("Welcome ", username)

	_, queue, err := pubsub.DeclareAndBind(
		rabbitConn,
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+username,
		routing.PauseKey,
		pubsub.SimpleQueueTransient,
	)
	if err != nil {
		log.Fatalf("could not subscribe to pause: %v", err)
	}
	fmt.Printf("Queue %v declared and bound!\n", queue.Name)
	// wait for ctrl+c
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	fmt.Println("RabbitMQ connection closed.")
}
