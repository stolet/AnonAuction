package common

import (
	"encoding/json"
	"errors"
	"time"
)

type AuctionStatus uint8

const (
	BEFORE AuctionStatus = iota
	DURING
	AFTER
)

type AuctionRound struct {
	Item         string
	StartTime    time.Time
	Interval     Duration
	Prices       []uint
	Auctioneers  []string
	T            int
	CurrentRound int  // If -1 it means the auction is over
}

type AwaitingCalculationMessage struct {
	CurrentRound int
}

type AuctionIsOverMessage struct {
	Message string
}

func (a *AuctionRound) AuctionStatus() AuctionStatus {
	if a.afterStartTime() && !a.afterEndTime() {
		return DURING
	} else if !a.afterEndTime() {
		return BEFORE
	} else {
		return AFTER
	}
}

func (a *AuctionRound) afterEndTime() bool {
	return time.Now().UTC().After(a.StartTime.Add(a.Interval.Duration))
}

func (a *AuctionRound) afterStartTime() bool {
	return time.Now().UTC().After(a.StartTime)
}

type Duration struct {
	time.Duration
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		d.Duration = time.Duration(value)
		return nil
	case string:
		var err error
		d.Duration, err = time.ParseDuration(value)
		if err != nil {
			return err
		}
		return nil
	default:
		return errors.New("invalid duration")
	}
}
