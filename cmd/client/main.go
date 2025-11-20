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
	// publishing channel
	publishCh, err := rabbitConn.Channel()
	if err != nil {
		log.Fatalf("could not open publish channel: %v", err)
	}
	// prompt for username
	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("error getting username: %v", err)
	}
	fmt.Println("Welcome ", username)
	// Create the game state for this client
	state := gamelogic.NewGameState(username)

	err = pubsub.SubscribeJSON(
		rabbitConn,
		routing.ExchangePerilDirect,
		routing.PauseKey+"."+state.GetUsername(),
		routing.PauseKey,
		pubsub.SimpleQueueTransient,
		true,
		handlerPause(state),
	)
	if err != nil {
		log.Fatalf("could not subscribe to pause: %v", err)
	}

	err = pubsub.SubscribeJSON(
		rabbitConn,
		routing.ExchangePerilTopic,
		routing.ArmyMovesPrefix+"."+state.GetUsername(),
		routing.ArmyMovesPrefix+".*",
		pubsub.SimpleQueueTransient,
		true,
		handlerMove(state),
	)
	if err != nil {
		log.Fatalf("could not subscribe to move queue: %v", err)
	}

	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}
		switch words[0] {
		case "spawn":
			err = state.CommandSpawn(words)
			if err != nil {
				fmt.Printf("error completing spawn command: %v\n", err)
				continue
			}
		case "move":
			mv, err := state.CommandMove(words)
			if err != nil {
				fmt.Printf("error completing move command: %v\n", err)
				continue
			}
			err = pubsub.PublishJSON(
				publishCh,
				routing.ExchangePerilTopic,
				routing.ArmyMovesPrefix+"."+mv.Player.Username,
				mv,
			)
			if err != nil {
				log.Printf("error: %s\n", err)
				continue
			}
			log.Printf("moved %v units to %v\n", len(mv.Units), len(mv.ToLocation))
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
