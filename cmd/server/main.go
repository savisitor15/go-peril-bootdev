package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/savisitor15/go-peril-bootdev/internal/pubsub"
	"github.com/savisitor15/go-peril-bootdev/internal/routing"
)

func main() {
	fmt.Println("Starting Peril server...")
	rabbitHost := "amqp://guest:guest@localhost:5672/"
	rabbitConn, err := amqp.Dial(rabbitHost)
	if err != nil {
		fmt.Println(fmt.Errorf("error creating connetion: %w", err))
	}
	defer rabbitConn.Close()
	fmt.Printf("connection esstablished to %s \n", rabbitHost)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	done := make(chan bool, 1)
	go func() {

		sig := <-sigs
		fmt.Println()
		fmt.Println(sig)
		done <- true
	}()
	// open a channel
	rabbitChan, err := rabbitConn.Channel()
	if err != nil {
		fmt.Println(fmt.Errorf("error creating connetion: %w", err))
	}
	err = pubsub.PublishJSON(rabbitChan, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true})
	if err != nil {
		fmt.Println(fmt.Errorf("error creating connetion: %w", err))
	}
	fmt.Println("awaiting signal")
	<-done
	fmt.Println("exiting")

}
