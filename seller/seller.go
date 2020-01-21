package seller

import (
	"../common"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"math"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const NOBID = "No Bid"
const MULTIPLEWINNERS = "Multiple Winners"

type Seller struct {
	AuctionRound common.AuctionRound
	router       *mux.Router
	publicKey    rsa.PublicKey
	privateKey   *rsa.PrivateKey
	// Key is Ip Port of auctioneer
	//BidPoints map[string]map[common.Price]common.Point

	agreedLagrangePoints map[common.Price]common.Point
	CloseAuction  bool
}

func Initialize(configFile string) *Seller {
	// Get configuration of the seller
	var auctionRound common.AuctionRound
	file, err := os.Open(configFile)
	defer file.Close()
	if err != nil {
		log.Fatalf("Error opening config file: %v", err)
		os.Exit(1)
	}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&auctionRound)
	if err != nil {
		log.Fatalf("Error decoding config file: %v", err)
		os.Exit(1)
	}

	// Some simple validations
	if auctionRound.T >= len(auctionRound.Auctioneers) {
		log.Fatalf("config file error: T value should be lower than length of auctioneers")
		os.Exit(1)
	}

	auctionRound.CurrentRound = 1

	// Create a new router
	rtr := mux.NewRouter()

	// Create a global seller
	privK, pubK := common.GenerateRSA() // Generate RSA key pair
	seller := &Seller{
		AuctionRound: auctionRound,
		router:       rtr,
		publicKey:    pubK,
		privateKey:   privK,
		//BidPoints:    make(map[string]map[common.Price]common.Point),
		agreedLagrangePoints: make(map[common.Price]common.Point),
		CloseAuction: false,
	}
	return seller
}

func (s *Seller) checkRoundTermination() {
	for {
		// Waiting for bidding round to end
		timeForEnd := time.Until(s.AuctionRound.StartTime.Add(s.AuctionRound.Interval.Duration))
		time.Sleep(timeForEnd)
		fmt.Println("Bidding phase is over. Waiting for lagrange calculation and all that stuff.")
		// Waiting for calculating round to end
		time.Sleep(s.AuctionRound.Interval.Duration / common.IntervalMultiple)


		for _, price := range s.AuctionRound.Prices {
			var individualLagrangePoints []common.Point
			for _, ipPort := range s.AuctionRound.Auctioneers {
				query := "http://" + ipPort + "/auctioneer/lagrange/" + strconv.FormatUint(uint64(price), 10)
				req, err := http.NewRequest("GET", query, nil)
				client := &http.Client{}
				var point common.Point

				resp, err := client.Do(req)
				if err != nil {
					//log.Println("error connecting to auctioneer, skipping... ")
					continue
				}
				defer resp.Body.Close()
				if err := json.NewDecoder(resp.Body).Decode(&point); err == nil {
					individualLagrangePoints = append(individualLagrangePoints, point)
				} else {
					fmt.Println("Didn't receive a valid point from auctioneer ", ipPort)
				}
			}
			// Need at least T auctioneers input?
			if len(individualLagrangePoints) <= s.AuctionRound.T {
					fmt.Println("We don't have more than T auctioneers contributing points. Ending auction.")
					time.Sleep(6 * time.Second)
					s.CloseAuction = true
					return
			}
			// Select the majority agreed upon lagrange value
			freqTable := make(map[string]int)
			var majority common.BigInt
			majorityOccurrences := 0
			for _, item := range individualLagrangePoints {
				if _, ok := freqTable[item.Y.Val.String()]; !ok {
					freqTable[item.Y.Val.String()] = 0
				}
				freqTable[item.Y.Val.String()] += 1
				if freqTable[item.Y.Val.String()] > majorityOccurrences {
					majorityOccurrences = freqTable[item.Y.Val.String()]
					majority = item.Y
				}
			}
			if  majorityOccurrences < (len(s.AuctionRound.Auctioneers) - s.AuctionRound.T) {
				fmt.Printf("Majority of auctioneers were not in agreement for price %v. Taking the value %v auctioneers agree on.\n", price, majorityOccurrences)
			} else {
				fmt.Printf("Majority of auctioneers were in agreement for price %v!\n", price)
			}

			s.agreedLagrangePoints[common.Price(price)] = common.Point{X:0, Y:majority}
		}

		//if len(s.BidPoints) < s.AuctionRound.T {
		//	fmt.Println("We have less than T auctioneers :(")
		//	time.Sleep(6 * time.Second)
		//	s.CloseAuction = true
		//	return
		//}
		//for _, priceMap := range s.agreedLagrangePoints {
		//	isDone := false
		//	for i := len(priceMap) - 1; i >= 0; i-- {
		//		price := common.Price(s.AuctionRound.Prices[i])
		//		encryptedID := priceMap[price]
		//		res := s.decodeID(encryptedID.Y.Val.Bytes())
		//		if res == NOBID {
		//			fmt.Println("There are no bids for price: ", price)
		//		} else if res == MULTIPLEWINNERS {
		//			fmt.Println("There are multiple winners for price: ", price)
		//			s.calculateNewRound(uint(price))
		//			isDone = true
		//			break
		//		} else {
		//			fmt.Println("Got a winner: ", res)
		//			s.contactWinner(res, price)
		//			s.AuctionRound.CurrentRound = -1
		//			time.Sleep(6 * time.Second)
		//			s.CloseAuction = true
		//			return
		//		}
		//	}
		//	if isDone == true {
		//		break
		//	}
		//}
		for i := len(s.AuctionRound.Prices) - 1; i >= 0; i-- {
			price := common.Price(s.AuctionRound.Prices[i])
			correspLPoint := s.agreedLagrangePoints[price]
			res := s.decodeID(correspLPoint.Y.Val.Bytes())
			if res == NOBID {
				fmt.Println("There are no bids for price: ", price)
			} else if res == MULTIPLEWINNERS {
				fmt.Println("There are multiple winners for price: ", price)
				s.calculateNewRound(uint(price))
				break
			} else {
				fmt.Println("Got a winner: ", res)
				s.contactWinner(res, price)
				s.AuctionRound.CurrentRound = -1
				time.Sleep(6 * time.Second)
				s.CloseAuction = true
				return
			}
		}
	}
}

func (s *Seller) StartAuction(address string) {
	s.router.HandleFunc("/seller/key", s.GetPublicKey).Methods("GET")
	s.router.HandleFunc("/seller/roundinfo", s.GetRoundInfo).Methods("GET")

	// Run the REST server
	go s.checkRoundTermination()
	log.Printf("Error: %v", http.ListenAndServe(address, s.router))
}

func (s *Seller) GetPublicKey(w http.ResponseWriter, r *http.Request) {
	data := common.MarshalKeyToPem(s.publicKey)
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Write(data)
}

func (s *Seller) GetRoundInfo(w http.ResponseWriter, r *http.Request) {
	data, err := json.Marshal(s.AuctionRound)
	if err != nil {
		log.Fatalf("error on GetRoundInfo: %v", err)
	}
	w.Write(data)
}

// Seller's private function ===========
func (s *Seller) decodeID(msg []byte) string {
	if len(msg) == 0 {
		return "No Bid"
	}
	// Attempt to decode the message.
	rawMsg, err := common.DecryptID(msg, s.privateKey)
	if err != nil {
		return "Multiple Winners"
	}
	return string(rawMsg)
}

func (s *Seller) contactWinner(ipPortAndPrice string, price common.Price) {
	ipPort := strings.Split(ipPortAndPrice, " ")[0]
	conn, err := net.Dial("tcp", ipPort)
	if err != nil {
		fmt.Println("Was not able to contact winning bidder: ", err)
        return
	}
	winnerNotification := common.WinnerNotification{WinningPrice: price}
	notifBytes, err := json.Marshal(winnerNotification)
	if err != nil {
		fmt.Println("Issue encoding winner notification: ", err)
	}
	conn.Write(notifBytes)
	conn.Close()
}

func (s *Seller) calculateNewRound(highestBid uint) {
	prices, _ := s.CalculateNewPrices(highestBid)
	newAuctionRound := common.AuctionRound{
		Item:         s.AuctionRound.Item,
		StartTime:    time.Now().Add(s.AuctionRound.Interval.Duration),
		Interval:     s.AuctionRound.Interval,
		Prices:       prices,
		Auctioneers:  s.AuctionRound.Auctioneers,
		T:            s.AuctionRound.T,
		CurrentRound: s.AuctionRound.CurrentRound + 1,
	}
	s.AuctionRound = newAuctionRound

}

func (s *Seller) CalculateNewPrices(highestBid uint) ([]uint, error) {
	numberOfPrices := len(s.AuctionRound.Prices)
	if numberOfPrices != 0 {
		priceInterval := s.AuctionRound.Prices[1] - s.AuctionRound.Prices[0]
		if s.AuctionRound.Prices[numberOfPrices-1] == highestBid {
			var newPrices []uint
			for i := 0; i < numberOfPrices; i++ {
				newPrices = append(newPrices, highestBid+uint(i)*priceInterval)
			}
			return newPrices, nil
		} else {
			newPriceInterval := uint(math.Ceil(float64(priceInterval) / float64(numberOfPrices)))
			var newPrices []uint
			for i := 0; i < numberOfPrices; i++ {
				newPrices = append(newPrices, uint(highestBid+uint(i)*newPriceInterval))
			}
			return newPrices, nil
		}
	} else {
		return nil, errors.New("Seller price list is empty!")
	}
}
