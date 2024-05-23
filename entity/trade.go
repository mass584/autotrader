package entity

import "time"

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
