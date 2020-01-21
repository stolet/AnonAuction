package main

import (
	"./auctioneer"
	"encoding/json"
	"log"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("Usage: auctioneer_main.go [initial_config_file_location]")
		os.Exit(1)
	}

	var config auctioneer.Config
	file, err := os.Open(os.Args[1])
	defer file.Close()
	if err != nil {
		log.Fatalf("Error opening config file: %v", err)
		os.Exit(1)
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatalf("Error decoding config file: %v", err)
		os.Exit(1)
	}

	auctioneer := auctioneer.Initialize(config)
	auctioneer.Start()
}
