package common

import (
	"math/big"
)

type lagrangePoint struct {
	X *big.Int
	Y *big.Int
}

type lagrangePoints []lagrangePoint

func (ps lagrangePoints) lagrange() *big.Int {
	result := new(big.Rat)
	lenPS := len(ps)

	for i := 0; i < lenPS; i++ {
		//Yterm := new(big.Float).SetInt(ps[i].Y)
		Yterm := new(big.Rat).SetInt(ps[i].Y)
		//Zterm := big.NewFloat(1)
		Zterm := new(big.Rat).SetInt64(int64(1))

		for j := 0; j < lenPS; j++ {
			if j != i {
				numo := new(big.Rat)
				numo.Sub(big.NewRat(int64(0), int64(1)), new(big.Rat).SetInt(ps[j].X))
				Yterm.Mul(Yterm, numo)

				deno := new(big.Rat)
				deno.Sub(new(big.Rat).SetInt(ps[i].X), new(big.Rat).SetInt(ps[j].X))
				Zterm.Mul(Zterm, deno)
			}
		}
		result.Add(result, new(big.Rat).Quo(Yterm, Zterm))
	}

    intResult := result.Num()

	return intResult
}

func ComputeLagrange(compressedPoints []CompressedPoints) map[Price]BigInt {
	lagrangeMap := make(map[Price]lagrangePoints)
	for _, cp := range compressedPoints {
		for k, v := range cp.Points {
			lagrangeMap[k] = append(lagrangeMap[k], lagrangePoint{
				X: big.NewInt(int64(v.X)),
				Y: big.NewInt(0).SetBytes(v.Y.Val.Bytes())}) // copy over the value
		}
	}

	interpolationMap := make(map[Price]BigInt)
	for k, v := range lagrangeMap {
		interpolationMap[k] = BigInt{v.lagrange()}
	}
	return interpolationMap
}
