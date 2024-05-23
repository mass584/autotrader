package entity

import "time"

type ExchangePlace int

// DBに永続化されるので順番を変えないこと
const (
	Bitflyer ExchangePlace = iota + 1
	Coincheck
)

type ExchangePair int

// DBに永続化されるので順番を変えないこと
const (
	BTC_TO_JPY ExchangePair = iota + 1
	ETH_TO_JPY
	ETH_TO_BTC
	ETC_TO_JPY
	XRP_TO_JPY
	BCH_TO_BTC
)

type Order struct {
	Price  float64
	Volume float64
}

type OrderBook struct {
	Asks []Order
	Bids []Order
}

type Trade struct {
	ID            int
	ExchangePlace ExchangePlace
	ExchangePair  ExchangePair
	TradeID       int
	Price         float64
	Volume        float64
	Time          time.Time
}

type TradeCollection []Trade

func (tc TradeCollection) RecentTrades(duration time.Duration) TradeCollection {
	cutoff := time.Now().Add(-duration)
	var filteredTrades TradeCollection
	for _, trade := range tc {
		if trade.Time.After(cutoff) {
			filteredTrades = append(filteredTrades, trade)
		}
	}
	return filteredTrades
}

func (tc TradeCollection) LatestTrade() Trade {
	// TradeCollectionはソート済み
	return tc[0]
}
