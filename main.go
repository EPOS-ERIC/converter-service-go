package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/epos-eu/converter-service/rabbit"
	"github.com/epos-eu/converter-service/server"
)

func main() {
	// context to handle shutdown gracefully
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	broker := rabbit.NewBroker()
	// start the broker handling
	err := broker.Start()
	if err != nil {
		panic(err)
	}
	// start to monitor the connection and automatically restart it (in place)
	go broker.Monitor(ctx)

	// start the server (blocking)
	server.StartServer(broker)
}
