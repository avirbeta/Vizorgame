package main

import (
	"log"

	"github.com/avirbeta/Vizorgame/game"
)

func main() {
	if err := game.RunDesktop(); err != nil {
		log.Fatal(err)
	}
}
