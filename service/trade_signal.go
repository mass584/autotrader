package service

import (
	"time"

	"github.com/mass584/autotrader/entity"
	"github.com/mass584/autotrader/repository/database"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type Decision string

const (
	Sell Decision = "SELL"
	Buy  Decision = "BUY"
	Hold Decision = "HOLD"
)

func CalculateTradeSignalOnCoincheck(db *gorm.DB, exchangePair entity.ExchangePair, signalAt time.Time) (Decision, Decision) {
	from := signalAt.Add(-50*24*time.Hour - 1*time.Minute)
	to := signalAt
	tradeCollection := database.GetTradesByTimeRange(db, entity.Coincheck, exchangePair, from, to)

	trendFollowSignal, _ := trendFollowingSignal(tradeCollection)
	log.Info().Msgf("TrandFollowSignal is: %s\n", trendFollowSignal)

	meanReversionSignal, _ := meanReversionSignal(tradeCollection)
	log.Info().Msgf("MeanReversionSignal is: %s\n", meanReversionSignal)

	// 一旦、トレンドフォローシグナルとミーンリバージョンシグナルを両方返しておく。
	// 実際は、これらの相関に応じて1つの決定値を返すように、何らかのルールを科す必要がある。
	return trendFollowSignal, meanReversionSignal
}

func calculateSimpleMovingAverage(trades entity.TradeCollection, period time.Duration) float64 {
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

func trendFollowingSignal(tradeCollection entity.TradeCollection) (Decision, float64) {
	// 一般的なパラメータとして、短期移動平均と長期移動平均の期間を10日と50日とする
	shortSMA := calculateSimpleMovingAverage(tradeCollection, 10*24*time.Hour)
	longSMA := calculateSimpleMovingAverage(tradeCollection, 50*24*time.Hour)
	currentPrice := tradeCollection[len(tradeCollection)-1].Price

	if shortSMA > longSMA {
		return Buy, currentPrice
	} else if shortSMA < longSMA {
		return Sell, currentPrice
	}
	return Hold, currentPrice
}

func meanReversionSignal(tradeCollection entity.TradeCollection) (Decision, float64) {
	// どれくらいの期間での単純移動平均を取るかのパラメータチューニングが必要
	sma := calculateSimpleMovingAverage(tradeCollection, 1*time.Minute)
	currentPrice := tradeCollection[len(tradeCollection)-1].Price

	if currentPrice < sma {
		return Buy, currentPrice
	} else if currentPrice > sma {
		return Sell, currentPrice
	}
	return Hold, currentPrice
}
