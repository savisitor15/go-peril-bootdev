package main

import (
	"fmt"
	"log"

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

	state := gamelogic.NewGameState(username)

	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}
		switch words[0] {
		case "spawn":
			err = state.CommandSpawn(words)
			if err != nil {
				fmt.Printf("error completing spawn command: %v", err)
				continue
			}
		case "move":
			army, err := state.CommandMove(words)
			if err != nil {
				fmt.Printf("error completing move command: %v", err)
			}
			log.Printf("Army of %v, with units %v moved to %v", army.Player.Username, army.Units[0].Rank, army.ToLocation)
		case "status":
			state.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "quit":
			gamelogic.PrintQuit()
			return
		default:
			fmt.Printf("unkown command")
		}
	}
}
