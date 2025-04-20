package main

import (
	"log"
	"os"
	"phase4/internal/app/config"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalln(err)
		os.Exit(1)
	}

	log.Printf("Debug: %v\n", cfg)
}
