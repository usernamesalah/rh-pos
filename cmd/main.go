package main

import (
	"log"

	"github.com/usernamesalah/rh-pos/internal/cli"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	cli.Execute()
}
