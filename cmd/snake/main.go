package main

import (
	"log"

	snake "github.com/isaporiti/snake"
)

func main() {
	if err := snake.Run(); err != nil {
		log.Fatal(err)
	}
}
