package auctioneer

import (
	//"fmt"
	"testing"
	"../common"
	"math/big"
)

func TestPermutations1(t *testing.T) {
	// 5 auctioneers, T=2
	// 4 prices: 20, 45, 70, 95
	// 2 bidders:
	// B1 polynomials (willing to pay 20 only, secretID 4000):
	// price 20: 5x^2 + 2x + 4000
	// price 45: 2x^2 + 3x + 0
	// price 70: x^2 + 8x + 0
	// price 95: 6x^2 + 2x + 0

	// B2 polynomials (willing to pay up to 70, secretID 3999):
	// price 20: 6x^2 + 2x + 3999
	// price 45: 3x^2 + 3x + 3999
	// price 70: 2x^2 + 8x + 3999
	// price 95: 7x^2 + 2x + 0
	T := 2
	var compressedPoints []common.CompressedPoints
	for i := 1; i < 6; i++ {
		Aimap := make(map[common.Price]common.Point)
		for _, price := range []uint{20,45,70,95} {
			var b1evalY, b2evalY int
			if price == 20 {
				b1evalY = 5*(i*i) + 2*i + 4000
				b2evalY = 6*(i*i) + 2*i + 3999
			} else if price == 45 {
				b1evalY = 2*(i*i) + 3*i + 0
				b2evalY = 3*(i*i) + 3*i + 3999
			} else if price == 70 {
				b1evalY = (i*i) + 8*i + 0
				b2evalY = 2*(i*i) + 8*i + 3999
			} else {
				b1evalY = 6*(i*i) + 2*i + 0
				b2evalY = 7*(i*i) + 2*i + 0
			}
			Aimap[common.Price(price)] = common.Point{X: i, Y: common.BigInt{big.NewInt(int64(b1evalY+b2evalY))}}
		}
		compressedPoints = append(compressedPoints, common.CompressedPoints{Aimap})
	}

	//t.Fatal(len(getPermutation(compressedPoints, T+1)))
	res := getPermutation(compressedPoints, T+1)
	t.Logf("%d groups: %v\n\n", len(res), res)

	for _, group := range res {
		res2 := common.ComputeLagrange(group)
		t.Logf("for group: %v got computed lagrange: %v\n", group, res2)
	}

	t.Fatal(".")
}

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
