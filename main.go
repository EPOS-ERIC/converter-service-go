package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/epos-eu/converter-service/db"
	"github.com/epos-eu/converter-service/rabbit"
	"github.com/epos-eu/converter-service/server"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := db.Init(); err != nil {
		panic("failed to connect to database: " + err.Error())
	}

	broker := rabbit.NewBroker()
	// start the broker handling
	err := broker.Start()
	if err != nil {
		panic(err)
	}
	// start to monitor the connection and automatically restart it (in place)
	go broker.Monitor(ctx)

	server.StartServer(broker)
}
