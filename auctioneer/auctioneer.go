package auctioneer

import (
	"../common"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"math/big"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type Auctioneer struct {
	config      Config
	round       uint
	bidMutex    *sync.Mutex
	currentBids map[common.Price][]common.Point
	roundInfo   common.AuctionRound
	lagrangeMap common.CompressedPoints // Calculated average lagrange value for seller to query
}

type Config struct {
	SellerIpPort   string
	LocalIpPort    string
	ExternalIpPort string
}

type AuctionRpcServer struct {
	Auctioneer *Auctioneer
}

func Initialize(config Config) *Auctioneer {
	return &Auctioneer{config: config,
		currentBids: make(map[common.Price][]common.Point),
		lagrangeMap: common.CompressedPoints{make(map[common.Price]common.Point)},
		bidMutex:    &sync.Mutex{}}
}

func (a *Auctioneer) Start() {
	a.UpdateRoundInfo()

	rtr := mux.NewRouter()
	rtr.HandleFunc("/auctioneer/sendBid", a.SendBid).Methods("POST")
	rtr.HandleFunc("/auctioneer/compressedPoints", a.GetCompressedPoints).Methods("GET")
	rtr.HandleFunc("/auctioneer/lagrange/{price:[0-9]+}", a.GetLagrange).Methods("GET")
	go a.runAuction()
	log.Println("Starting the auctioneer server...")
	err := http.ListenAndServe(a.config.LocalIpPort, rtr)
	log.Printf("Error: %v", err)
}

func (a *Auctioneer) UpdateRoundInfo() {
	currentRoundNumber := a.roundInfo.CurrentRound
	req, err := http.NewRequest("GET", "http://"+a.config.SellerIpPort+"/seller/roundinfo", nil)
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Do: ", err)
		return
	}

	defer resp.Body.Close()

	if err != nil {
		log.Fatal("NewRequest: ", err)
	}
	var roundInfo common.AuctionRound
	if err := json.NewDecoder(resp.Body).Decode(&roundInfo); err != nil {
		log.Println(err)
	}
	a.roundInfo = roundInfo
	if currentRoundNumber+1 == a.roundInfo.CurrentRound {
		a.currentBids = make(map[common.Price][]common.Point)
	}
}

// Receives bids from a bidder and returns if true if it was successfully received
func (a *Auctioneer) SendBid(w http.ResponseWriter, r *http.Request) {
	status := a.roundInfo.AuctionStatus()
	if status == common.DURING {

		var bidPoints common.BidPoints
		err := json.NewDecoder(r.Body).Decode(&bidPoints)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("Received bid from ", bidPoints.BidderID)
		a.bidMutex.Lock()
		for price, points := range bidPoints.Points {
			a.currentBids[price] = append(a.currentBids[price], points)
		}
		a.bidMutex.Unlock()
		w.WriteHeader(200)
	} else {
		w.WriteHeader(503)
	}
}

func (a *Auctioneer) GetCompressedPoints(w http.ResponseWriter, r *http.Request) {
	a.bidMutex.Lock()
	defer a.bidMutex.Unlock()
	err := json.NewEncoder(w).Encode(a.calculateCompressedPoints())
	if err != nil {
		fmt.Println(err)
	}
}

func (a *Auctioneer) calculateCompressedPoints() common.CompressedPoints {
	compressedPoints := common.CompressedPoints{make(map[common.Price]common.Point)}
	for key, points := range a.currentBids {
		var sum big.Int
		for _, p := range points {
			point := p.Y.Val
			sum.Add(&sum, point)
		}
		compressedPoints.Points[key] = common.Point{points[0].X, common.BigInt{&sum}}
	}

	return compressedPoints
}

//Meant to run as a go routine
func (a *Auctioneer) runAuction() {
	round := 0
	for {
		for {
			a.UpdateRoundInfo()
			if a.roundInfo.CurrentRound > round {
				round = a.roundInfo.CurrentRound
				break
			} else if a.roundInfo.CurrentRound == -1 {
				fmt.Println("Auction over")
				return
			}
			// some backoff for querying
			time.Sleep(time.Second)
		}
		a.runAuctionRound()
	}
}

func (a *Auctioneer) runAuctionRound() {
	if a.roundInfo.AuctionStatus() == common.AFTER {
		fmt.Println("Auction over")
		return
	}

	until := time.Until(a.roundInfo.StartTime)
	fmt.Println("Waiting for ", until, " before auction")
	time.Sleep(until)
	fmt.Println("Auction started")
	time.Sleep(a.roundInfo.Interval.Duration)
	fmt.Println("Collecting compressed points")

	a.bidMutex.Lock()
	compressedPoints := []common.CompressedPoints{a.calculateCompressedPoints()}
	a.bidMutex.Unlock()

	for _, ipPort := range a.roundInfo.Auctioneers {
		if ipPort != a.config.ExternalIpPort {
			points, err := a.QueryCompressed(ipPort)
			if err == nil {
				compressedPoints = append(compressedPoints, points)
			} else {
				fmt.Println(err)
			}
		}
	}

	// Get T+1 group permutation from other auctioneer's point
	groupPermutation := getPermutation(compressedPoints, a.roundInfo.T+1)
	fmt.Printf("Will be searching over %v permutations of subsets.\n", len(groupPermutation))

	// Calculate lagrange for each permutation group and keep track of majority
	freqLagrange := make(map[common.Price]map[string]int)
	for _, group := range groupPermutation {
		//fmt.Println("Calculating for group: ", group)
		res := common.ComputeLagrange(group)
		for k, v := range res {
			if _, ok := freqLagrange[k]; !ok {
				freqLagrange[k] = make(map[string]int)
				freqLagrange[k][v.Val.String()] = 0
			}
			freqLagrange[k][v.Val.String()] += 1
		}
	}
	// Calculate majority from all lagrange permutation we calculated
	a.lagrangeMap = common.CompressedPoints{make(map[common.Price]common.Point)}
	for price, bidMap := range freqLagrange {
		currCount := 0
		currID := big.NewInt(0)
		for id, frequency := range bidMap {
			if frequency > currCount {
				//currID = id.Val
				newI, _  := big.NewInt(0).SetString(id, 10)
				currID = newI
				// Added following
				currCount = frequency
			}
		}
		fmt.Printf("%v subsets of size T+1 were in agreement on the Lagrange value.\n", currCount) 
		a.lagrangeMap.Points[price] = common.Point{Y: common.BigInt{currID}}
	}
	time.Sleep(time.Until(a.roundInfo.StartTime.Add(a.roundInfo.Interval.Duration).Add(a.roundInfo.Interval.Duration / common.IntervalMultiple)))
}

func (a *Auctioneer) GetLagrange(w http.ResponseWriter, r *http.Request) {
	price := mux.Vars(r)["price"]
	uPrice, err := strconv.ParseUint(price, 10, 32)
	fmt.Println("Got lagrange request from seller for price ", price)
	if err != nil {
		fmt.Println("Error parsing into uint: ", err)
		return
	}
	payload := a.lagrangeMap.Points[common.Price(uPrice)]
	data, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("error on marshalling: ", err)
	}
	w.Write(data)
}

func (a *Auctioneer) QueryCompressed(ipPort string) (common.CompressedPoints, error) {
	fmt.Println("Getting compressed point from ", ipPort)
	req, err := http.NewRequest("GET", "http://"+ipPort+"/auctioneer/compressedPoints", nil)
	client := &http.Client{}
	var compressedPoints common.CompressedPoints

	resp, err := client.Do(req)
	if err != nil {
		return compressedPoints, err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&compressedPoints); err != nil {
		return compressedPoints, err
	}

	return compressedPoints, nil
}

func (a *Auctioneer) SendTotalPoints(bids common.TotalBids) {
	jsonBytes, e := json.Marshal(bids)
	if e != nil {
		panic(e)
	}
	req, err := http.NewRequest("POST", "http://"+a.config.SellerIpPort+"/seller/bidpoint", bytes.NewBuffer(jsonBytes))
	client := &http.Client{}

	_, e = client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
}

// Given a whole list of compressed points, it returns a list of list with permutation of tValue
func getPermutation(compressedPoints []common.CompressedPoints, group int) [][]common.CompressedPoints {
	res := make([][]common.CompressedPoints, 0)
	if len(compressedPoints) < group {
		return res
	}

	if group == 1 {
		for _, cp := range compressedPoints {
			res = append(res, []common.CompressedPoints{cp})
		}
		return res
	}

	for i, cp := range compressedPoints {
		childRes := getPermutation(compressedPoints[i+1:], group-1)
		for _, childList := range childRes {
			res = append(res, append([]common.CompressedPoints{cp}, childList...))
		}
	}
	return res
}
