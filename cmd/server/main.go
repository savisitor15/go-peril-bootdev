package main

import (
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/savisitor15/go-peril-bootdev/internal/gamelogic"
	"github.com/savisitor15/go-peril-bootdev/internal/pubsub"
	"github.com/savisitor15/go-peril-bootdev/internal/routing"
)

func main() {
	log.Println("Starting Peril server...")
	rabbitHost := "amqp://guest:guest@localhost:5672/"
	rabbitConn, err := amqp.Dial(rabbitHost)
	if err != nil {
		log.Fatalf("error creating connetion: %v", err)
	}
	defer rabbitConn.Close()
	log.Printf("connection esstablished to %s \n", rabbitHost)
	// open a channel
	rabbitChan, err := rabbitConn.Channel()
	if err != nil {
		log.Fatalf("error creating connetion: %v", err)
	}
	// log queue declare and bind
	_, logqueue, err := pubsub.DeclareAndBind(
		rabbitConn,
		routing.ExchangePerilTopic,
		routing.GameLogSlug,
		routing.GameLogSlug+".*",
		pubsub.SimpleQueueDurable,
		false)
	if err != nil {
		log.Fatalf("error binding to game_logs queue: %v", err)
	}
	log.Printf("Queue %v declared and bound.", logqueue.Name)

	// Print help
	gamelogic.PrintServerHelp()

	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}
		switch words[0] {
		case "pause":
			fmt.Println("Publishing paused game state")
			err = pubsub.PublishJSON(
				rabbitChan,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{
					IsPaused: true,
				},
			)
			if err != nil {
				log.Printf("could not publish pause state: %v", err)
			}
		case "resume":
			fmt.Println("publishing resumes game state")
			err = pubsub.PublishJSON(
				rabbitChan,
				routing.ExchangePerilDirect,
				routing.PauseKey,
				routing.PlayingState{
					IsPaused: false,
				},
			)
			if err != nil {
				log.Printf("could not publish pause state: %v", err)
			}
		case "quit":
			log.Println("goodbye")
			return
		default:
			fmt.Printf("unkown command")
		}
	}
}
