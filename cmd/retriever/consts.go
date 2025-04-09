package main

import (
	"crypto/rand"
	"math/big"
)

const (
	// baseURL is the base URL for the API
	baseURL = "https://www.englandgolf.org/api"

	// maxTickerDuration is the maximum duration for the ticker
	maxTickerDurationSec = 180

	// minTickerDuration is the minimum duration for the ticker
	minTickerDurationSec = 60
)

func ticketMagicNumber() int64 {
	randInt, err := rand.Int(rand.Reader, big.NewInt(maxTickerDurationSec-minTickerDurationSec+1))
	if err != nil {
		return minTickerDurationSec
	}

	return minTickerDurationSec + randInt.Int64()
}
