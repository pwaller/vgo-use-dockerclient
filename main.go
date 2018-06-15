package main

import (
	"log"

	"github.com/docker/docker/client"
)

func main() {
	c, err := client.NewEnvClient()
	log.Println(c, err)
}
