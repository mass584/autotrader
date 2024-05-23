package trade_algorithms

import (
	"time"

	"github.com/mass584/autotrader/entity"
)

type Decision string

const (
	Sell Decision = "SELL"
	Buy  Decision = "BUY"
	Hold Decision = "HOLD"
)

func CalculateSimpleMovingAverage(trades entity.TradeCollection, period time.Duration) float64 {
	cutoff := trades.LatestTrade().Time.Add(-period)

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

func TrendFollowingSignal(tradeCollection entity.TradeCollection) (Decision, float64) {
	// どれくらいの期間での単純移動平均を取るかのパラメータチューニングが必要
	shortSMA := CalculateSimpleMovingAverage(tradeCollection, 5*time.Minute)
	longSMA := CalculateSimpleMovingAverage(tradeCollection, 50*time.Minute)
	currentPrice := tradeCollection[len(tradeCollection)-1].Price

	if shortSMA > longSMA {
		return Buy, currentPrice
	} else if shortSMA < longSMA {
		return Sell, currentPrice
	}
	return Hold, currentPrice
}

func MeanReversionSignal(tradeCollection entity.TradeCollection) (Decision, float64) {
	// どれくらいの期間での単純移動平均を取るかのパラメータチューニングが必要
	sma := CalculateSimpleMovingAverage(tradeCollection, 1*time.Minute)
	currentPrice := tradeCollection[len(tradeCollection)-1].Price

	if currentPrice < sma {
		return Buy, currentPrice
	} else if currentPrice > sma {
		return Sell, currentPrice
	}
	return Hold, currentPrice
}
