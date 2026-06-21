package main

import (
	"log"

	"github.com/RaghtNow/opc/apps/edu-insight-api/internal/app"
)

func main() {
	server, err := app.NewServer()
	if err != nil {
		log.Fatal(err)
	}
	defer server.Close()

	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
