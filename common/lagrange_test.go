package common

import (
	"github.com/jongukim/polynomial"
	"log"
	"math/big"
	"testing"
)

func TestHelloWorld(t *testing.T) {
	// Step 0) Generate RSA for seller
	privateK, publicK := GenerateRSA()

	// T_val will be 1 and our auctioneers will be 2
	T_VAL := 1

	// ==== Bidder 1
	// Step 1.1) Bidder1 creates an ID
	bidder1RawID, _ := EncryptID("192.168.111.231:9999", 200, &publicK)
	bidder1ID := big.NewInt(0)

	// Step 1.2) Generate a polynomial, we assume T value is 1 and number of auctioneers are 2
	bidder1Poly100 := generatePolynomial(T_VAL, big.NewInt(0))
	bidder1Poly200 := generatePolynomial(T_VAL, bidder1ID.SetBytes(bidder1RawID))

	// Step 1.3) Bidder1 sends X=1 and X=2 to auctioneer 1 and 2 respectively
	// To auctioneer 1
	bidder1Auctioneer1Price100 := Point{
		X: 1,
		Y: BigInt{
			bidder1Poly100.Eval(big.NewInt(1), nil),
		},
	}
	bidder1Auctioneer1Price200 := Point{
		X: 1,
		Y: BigInt{
			bidder1Poly200.Eval(big.NewInt(1), nil),
		},
	}

	// To auctioneer 2
	bidder1Auctioneer2Price100 := Point{
		X: 2,
		Y: BigInt{
			bidder1Poly100.Eval(big.NewInt(2), nil),
		},
	}
	bidder1Auctioneer2Price200 := Point{
		X: 2,
		Y: BigInt{
			bidder1Poly200.Eval(big.NewInt(2), nil),
		},
	}

	// ==== Bidder 2
	// Step 1.3) Bidder2 creates an ID
	bidder2RawID, _ := EncryptID("201:221:3133:9999", 100, &publicK)
	bidder2ID := big.NewInt(0)

	// Step 1.4) Generate a polynomial, we assume T value is 1 and number of auctioneers are 2
	bidder2Poly100 := generatePolynomial(T_VAL, bidder2ID.SetBytes(bidder2RawID))
	//bidder2Poly100 := generatePolynomial(T_VAL, big.NewInt(12343))
	bidder2Poly200 := generatePolynomial(T_VAL, big.NewInt(0))

	// Step 1.3) Bidder1 sends X=1 and X=2 to auctioneer 1 and 2 respectively
	// To auctioneer 1
	bidder2Auctioneer1Price100 := Point{
		X: 1,
		Y: BigInt{
			bidder2Poly100.Eval(big.NewInt(1), nil),
		},
	}
	bidder2Auctioneer1Price200 := Point{
		X: 1,
		Y: BigInt{
			bidder2Poly200.Eval(big.NewInt(1), nil),
		},
	}
	// To auctioneer 2
	bidder2Auctioneer2Price100 := Point{
		X: 2,
		Y: BigInt{
			bidder2Poly100.Eval(big.NewInt(2), nil),
		},
	}
	bidder2Auctioneer2Price200 := Point{
		X: 2,
		Y: BigInt{
			bidder2Poly200.Eval(big.NewInt(2), nil),
		},
	}

	// Step 2) auctioneer 1 and 2 compress their points and send it to each other. They SHOULD have the same compressed points.
	//          We will be mocking the compression by manually adding two points

	// Auctioneer 1 compresses its values for 100 and 200 price range
	var auctioneer1Compressed CompressedPoints
	auctioneer1Compressed.Points = make(map[Price]Point)
	auctioneer1Compressed.Points[100] = Point{
		X: bidder1Auctioneer1Price100.X,
		Y: BigInt{
			Val: big.NewInt(0).Add(bidder1Auctioneer1Price100.Y.Val, bidder2Auctioneer1Price100.Y.Val),
		},
	}
	auctioneer1Compressed.Points[200] = Point{
		X: bidder1Auctioneer1Price200.X,
		Y: BigInt{
			Val: big.NewInt(0).Add(bidder1Auctioneer1Price200.Y.Val, bidder2Auctioneer1Price200.Y.Val),
		},
	}

	// Auctioneer 2 compresses its values for 100 and 200 price range
	var auctioneer2Compressed CompressedPoints
	auctioneer2Compressed.Points = make(map[Price]Point)
	auctioneer2Compressed.Points[100] = Point{
		X: bidder1Auctioneer2Price100.X,
		Y: BigInt{
			Val: big.NewInt(0).Add(bidder1Auctioneer2Price100.Y.Val, bidder2Auctioneer2Price100.Y.Val),
		},
	}
	auctioneer2Compressed.Points[200] = Point{
		X: bidder1Auctioneer2Price200.X,
		Y: BigInt{
			Val: big.NewInt(0).Add(bidder1Auctioneer2Price200.Y.Val, bidder2Auctioneer2Price200.Y.Val),
		},
	}

	// Step 3) The above compressed info will be exchanged with each othher. Both acutioneers SHOULD have the same CompressedPoints result
	// An example of what an auctioneer would do
	var cps []CompressedPoints
	cps = append(cps, auctioneer1Compressed) // compressed poitns of auctioneer 1
	cps = append(cps, auctioneer2Compressed) // compressed points of auctioneer 2

	// Compute lagrange. The result will be sent over th the seller
	res := ComputeLagrange(cps)

	// Step 4) Seller decrypts for each price range using its private key
	for k, v := range res {
		rawID, err := DecryptID(v.Val.Bytes(), privateK)
		if err != nil {
			log.Printf("Error on decrypt: %v", err)
			continue
		}
		log.Printf("price: %v, decrypted id: %v", k, string(rawID))
	}
}

// f(x) = 3x^3 + 2x + 1 => [1 2 0 3]
func generatePolynomial(degree int, id *big.Int) polynomial.Poly {
	poly := polynomial.RandomPoly(int64(degree), 5) // 5 is hard coded to make coefficients 2^5 at most

	// Change the ID
	if id != nil {
		poly[0] = id
	} else {
		poly[0] = big.NewInt(0)
	}
	return poly
}
