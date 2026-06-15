package main

import (
	"log"

	"github.com/RaghtNow/opc/apps/edu-insight-api/internal/app"
)

func main() {
	server, err := app.NewServer()
	if err != nil {
		log.Fatalf("bootstrap server failed: %v", err)
	}

	if err := server.Run(); err != nil {
		log.Fatalf("server stopped with error: %v", err)
	}
}
