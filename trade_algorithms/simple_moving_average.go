package trade_algorithms

import (
	"time"

	"github.com/mass584/auto-trade/entity"
)

func CalculateSimpleMovingAverage(trades entity.TradeCollection, period time.Duration) float64 {
	cutoff := time.Now().Add(-period)

	var filteredTrades entity.TradeCollection
	for _, trade := range trades {
		if trade.Time.After(cutoff) {
			filteredTrades = append(filteredTrades, trade)
		}
	}

	var priceSum float64
	for _, trade := range filteredTrades {
		priceSum += trade.Price
	}

	return priceSum / float64(len(filteredTrades))
}
