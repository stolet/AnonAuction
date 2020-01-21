package main

import (
	"./seller"
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	log.Println("Starting seller client")
	if len(os.Args) != 3 {
		log.Fatalf("Usage: seller_main.go [REST_address (IP:PORT)] [initial_config_file_location]")
		os.Exit(1)
	}
	s := seller.Initialize(os.Args[2])
	go s.StartAuction(os.Args[1])
	log.Println("Started seller REST API")

	// Start the auction CLI
	if time.Now().UTC().After(s.AuctionRound.StartTime) {
		log.Printf("start time: %v", s.AuctionRound.StartTime)
		log.Printf("now: %v", time.Now().UTC())
		fmt.Println("Invalid start time.")
		return
	}

	fmt.Println("\n\n=====Starting the auction!=====")

	reader := bufio.NewReader(os.Stdin)

	var initialPrice int
	var err error
	for {
		fmt.Printf("Enter an initial price for the auction: ")
		initialPriceString, _ := reader.ReadString('\n')
		initialPrice, err = strconv.Atoi(strings.Replace(initialPriceString, "\n", "", 1))
		if err != nil {
			fmt.Println("Your initial price must be a number")
		} else {
			break
		}
	}
	var numPrices int
	for {
		fmt.Printf("Enter the number of prices for the auction: ")
		numPricesString, _ := reader.ReadString('\n')
		numPrices, err = strconv.Atoi(strings.Replace(numPricesString, "\n", "", 1))
		if err != nil {
			fmt.Println("The number of prices for the auction must be a number")
		} else {
			break
		}
	}
	var stride int
	for {
		fmt.Printf("Enter a stride for the prices: ")
		strideString, _ := reader.ReadString('\n')
		stride, err = strconv.Atoi(strings.Replace(strideString, "\n", "", 1))
		if err != nil {
			fmt.Println("The stride for the prices must be a number")
		} else {
			break
		}
	}

	var prices = []uint{}
	for i := 0; i < numPrices; i++ {
		prices = append(prices, uint(initialPrice+i*stride))
	}
	log.Println("Got pricerange: ", prices, " of length: ", len(prices))

	s.AuctionRound.Prices = prices
	s.AuctionRound.CurrentRound += 1

	fmt.Println("Waiting for current round to finish...")
	for {
		time.Sleep(time.Second) // Sleep and check every 1 second
		if s.CloseAuction {
			break
		}
	}
}
