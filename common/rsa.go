package common

/*
A simple helper class that deals all the RSA encryption.
*/

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/asn1"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
)

// Create RSA key pair in PEM format - used by seller only
func GenerateRSA() (*rsa.PrivateKey, rsa.PublicKey) {
	reader := rand.Reader
	bitSize := 2048
	key, err := rsa.GenerateKey(reader, bitSize)
	if err != nil {
		log.Fatalf("Error generating key: %v", err)
		// TODO: handle error
	}
	return key, key.PublicKey
}

// Marshal rsa public key
func MarshalKeyToPem(key rsa.PublicKey) []byte {
	asn1Bytes, err := asn1.Marshal(key)
	if err != nil {
		log.Fatalf("Error on encoding to pem: %v")
		// TODO: handle error
	}
	return pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: asn1Bytes,
	})
}

// Unmarshal rsa public key
func UnmarshalPemToKey(rawKey []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(rawKey)
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, errors.New("Error decoding the key")
	}

	var pk rsa.PublicKey
	_, err := asn1.Unmarshal(block.Bytes, &pk)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error on decoding pk: %v", pk))
	}

	return &pk, nil
}

// TODO: Return big.Int?
func EncryptID(ipAddressPort string, price uint, publicKey *rsa.PublicKey) ([]byte, error) {
	// ID will be bidder's IP address + price encrypted in seller's public key
	// EX: "127.0.0.1:9091 300" -> encrypted with public key
	// We use OAEP encryption standrad and NOT PKCK1
	plainByte := []byte(fmt.Sprintf("%v %v", ipAddressPort, price))
	rng := rand.Reader
	return rsa.EncryptOAEP(sha256.New(), rng, publicKey, plainByte, []byte(""))
}

func DecryptID(rawMsg []byte, privateKey *rsa.PrivateKey) ([]byte, error) {
	rng := rand.Reader
	return rsa.DecryptOAEP(sha256.New(), rng, privateKey, rawMsg, []byte(""))
}
