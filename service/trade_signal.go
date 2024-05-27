package service

import (
	"errors"
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
	// 過去50日分の取引データを取得する
	from := signalAt.Add(-50*24*time.Hour - 1*time.Minute)
	to := signalAt
	tradeCollection := database.GetTradesByTimeRange(db, entity.Coincheck, exchangePair, from, to)

	trendFollowSignal := trendFollowingSignal(tradeCollection, signalAt)
	log.Info().Msgf("TrandFollowSignal is: %s\n", trendFollowSignal)

	meanReversionSignal := meanReversionSignal(tradeCollection, signalAt)
	log.Info().Msgf("MeanReversionSignal is: %s\n", meanReversionSignal)

	// 一旦、トレンドフォローシグナルとミーンリバージョンシグナルを両方返しておく。
	// 実際は、これらの相関に応じて1つの決定値を返すように、何らかのルールを科す必要がある。
	return trendFollowSignal, meanReversionSignal
}

func calculateSimpleMovingAverage(trades entity.TradeCollection, signalAt time.Time, period time.Duration) (float64, error) {
	cutoff := signalAt.Add(-period)

	var filteredTrades entity.TradeCollection
	for _, trade := range trades {
		if trade.Time.After(cutoff) {
			filteredTrades = append(filteredTrades, trade)
		}
	}
	if len(filteredTrades) == 0 {
		return 0, errors.New("No trades in the period")
	}

	var priceSum float64
	for _, trade := range filteredTrades {
		priceSum += trade.Price
	}

	return priceSum / float64(len(filteredTrades)), nil
}

// 一般的なパラメータとして、短期移動平均と長期移動平均の期間を10日と50日とする
func trendFollowingSignal(tradeCollection entity.TradeCollection, signalAt time.Time) Decision {
	shortSMA, error := calculateSimpleMovingAverage(tradeCollection, signalAt, 10*24*time.Hour)
	if error != nil {
		return Hold
	}
	longSMA, error := calculateSimpleMovingAverage(tradeCollection, signalAt, 50*24*time.Hour)
	if error != nil {
		return Hold
	}

	if shortSMA > longSMA {
		return Buy
	} else if shortSMA < longSMA {
		return Sell
	}
	return Hold
}

// どれくらいの期間での単純移動平均を取るかのパラメータチューニングが必要
func meanReversionSignal(tradeCollection entity.TradeCollection, signalAt time.Time) Decision {
	sma, error := calculateSimpleMovingAverage(tradeCollection, signalAt, 10*time.Minute)
	if error != nil {
		return Hold
	}

	currentPrice := tradeCollection.LatestTrade().Price

	if currentPrice < sma {
		return Buy
	} else if currentPrice > sma {
		return Sell
	}
	return Hold
}
