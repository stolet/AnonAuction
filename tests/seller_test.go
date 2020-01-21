package tests

import (
	"../common"
	"../seller"
	"encoding/json"
	"net/http"
	"reflect"
	"testing"
	"time"
)

func TestGetConfig(t *testing.T) {
	s := seller.Initialize("../tests/test_config.json")
	go s.StartAuction("127.0.0.1:8787")
	time.Sleep(1 * time.Second)
	resp, _ := http.Get("http://localhost:8787/seller/roundinfo")
	var auctionRound common.AuctionRound
	json.NewDecoder(resp.Body).Decode(&auctionRound)
	if auctionRound.Item != "Fancy chocolate" {
		t.Errorf("Item is incorrect, got: %d, want: %d.", auctionRound.Item, "Fancy chocolate")
	}
	if !reflect.DeepEqual(auctionRound.Prices, []uint{300, 400, 500})  {
		t.Errorf("Prices are incorrect, got: %d, want: %d.", auctionRound.Prices, []uint{300, 400, 500})
	}
	if !reflect.DeepEqual(auctionRound.Auctioneers, []string{"127.0.0.1:8081", "127.0.0.1:8082"})  {
		t.Errorf("Auctioneers are incorrect, got: %d, want: %d.", auctionRound.Auctioneers, []string{"127.0.0.1:8081", "127.0.0.1:8082"})
	}
	if auctionRound.Interval.String() != "30s" {
		t.Errorf("Interval is incorrect, got: %d, want: %d.", auctionRound.Interval, "30s")
	}
	if auctionRound.T != 2 {
		t.Errorf("T_Value is incorrect, got: %d, want: %d.", auctionRound.T, 2)
	}
	if auctionRound.CurrentRound != 1 {
		t.Errorf("CurrRound is incorrect, got: %d, want: %d.", auctionRound.CurrentRound, 1)
	}
	expectedTime, _ := time.Parse("2006-01-02T15:04:05Z07:00","2018-12-01T16:00:00Z")
	if auctionRound.StartTime.String() != expectedTime.String() {
		t.Errorf("StartTime is incorrect, got: %d, want: %d.", auctionRound.StartTime, expectedTime)
	}

}

func TestNewPrices(t *testing.T) {
	s := seller.Initialize("test_config.json")
	newPrices, _ := s.CalculateNewPrices(400)
	expectedPrices := []uint{400, 434, 468}
	if !reflect.DeepEqual(newPrices, expectedPrices) {
		t.Errorf("Item was incorrect, got: %d, want: %d.", newPrices, expectedPrices)
	}

	newPrices, _ = s.CalculateNewPrices(500)
	expectedPrices = []uint{500, 600, 700}

}
