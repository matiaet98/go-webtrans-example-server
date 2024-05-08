package main

import (
	"context"
	"log"

	"github.com/matiaet98/go-webtrans-example-server/cmd"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	server := cmd.New(ctx)
	if err := server.Run(); err != nil {
		log.Fatal("Error running server: %w", err.Error())
		cancel()
	}
}
