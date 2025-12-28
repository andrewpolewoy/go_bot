package main

import (
	"log"

	"github.com/andrewpolewoy/go_bot/cmd/bot/internal/app"
)

func main() {
	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
