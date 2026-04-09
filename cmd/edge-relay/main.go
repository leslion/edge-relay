package main

import (
	"log"

	"github.com/leslion/edge-relay/internal"
)

func main() {
	if err := internal.Run(); err != nil {
		log.Fatal(err)
	}
}
