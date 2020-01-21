package common

import (
	"encoding/json"
	"math/big"
)

// this is the amount the interval is diveded by for the information exhange peroid
const IntervalMultiple  = 2

type Price uint

type BigInt struct {
	Val *big.Int
}

// X is int, since auctioneers are ordered {1,2,...,N} for small N
// Y is big.Int, since our IDs can get huge
type Point struct {
	X int
	Y BigInt
}

type BidPoints struct {
	BidderID string
	Points   map[Price]Point
}

type ListCompressedPoints struct {
	ListPoints []CompressedPoints
}

type CompressedPoints struct {
	Points map[Price]Point
}

type WinnerNotification struct {
	WinningPrice Price
}

type TotalBids struct {
	AuctioneerId string
	Points       CompressedPoints
}

func MarshalBidPoints(points BidPoints) ([]byte, error) {
	return json.Marshal(points)
}

func UnmarshalBidPoints(bidPointsBytes []byte, points *BidPoints) error {
	err := json.Unmarshal(bidPointsBytes, points)
	return err
}

func (bigint BigInt) MarshalJSON() ([]byte, error) {
	return []byte(bigint.Val.String()), nil
}

func (bigint *BigInt) UnmarshalJSON(b []byte) error {
	val := string(b[:])

	n := new(big.Int)
	n.SetString(val, 10)
	bigint.Val = n
	return nil
}
