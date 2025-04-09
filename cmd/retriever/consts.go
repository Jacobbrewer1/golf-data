package main

const (
	// baseURL is the base URL for the API
	baseURL = "https://www.englandgolf.org/api"

	// maxTickerDuration is the maximum duration for the ticker
	maxTickerDurationSec = 180

	// minTickerDuration is the minimum duration for the ticker
	minTickerDurationSec = 60
)

func ticketMagicNumber() int64 {
	return maxTickerDurationSec - minTickerDurationSec
}
