package main

import (
    "./bidder"
    "log"
    "os"
    "fmt"
	"bufio"
	"strconv"
	"strings"
	"net"
	"time"
)

// Hack to get current IP address
func thisIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP
}

func IntInList(list []uint, element uint) bool {
	for _, number := range list {
		if element == number {
			return true
		}
	}
	return false
}

func main() {
    //log.Println("Bidder client starting.")
    if len(os.Args) != 2 {
        log.Fatalf("Usage: bidder_client.go [seller_ip_address]")
        os.Exit(1)
    }

    // Initialize bidder
    bidder := bidder.InitBidder(os.Args[1], thisIP().String())

    for {
		bidder.LearnAuctionRound()
		fmt.Printf("The seller is selling \"%v\" at the following prices: %v.\n", bidder.RoundInfo.Item, bidder.RoundInfo.Prices)
		currentRound := bidder.RoundInfo.CurrentRound
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter your maximum bid: ")
		text, _ := reader.ReadString('\n')

		maxBid, err := strconv.Atoi(strings.TrimSpace(text))
		if err != nil {
			fmt.Println("Your bid was not understood: ", err)
			os.Exit(1)
		}
		fmt.Println(bidder.RoundInfo.Prices)
		if !IntInList(bidder.RoundInfo.Prices, uint(maxBid)) {
			fmt.Println("Your bid is not in the acceptable price range.")
			continue
		}
		bidder.ProcessBid(maxBid)
		// Send bids only once auction starts
		timeForStart := time.Until(bidder.RoundInfo.StartTime)
		time.Sleep(timeForStart)
		bidder.SendPoints()

		go bidder.ListenSeller()
		for {
			bidder.LearnAuctionRound()
			if bidder.RoundInfo.CurrentRound == -1 {
				fmt.Println("You've lost the auction.")
				return
			} else if bidder.RoundInfo.CurrentRound == currentRound + 1 {
				timeForStart := time.Until(bidder.RoundInfo.StartTime)
				fmt.Println("Auction was tied. Going to tie-breaking round in ", timeForStart)
				time.Sleep(timeForStart)
				break
			}
			time.Sleep(3 * time.Second)
		}
	}
}
