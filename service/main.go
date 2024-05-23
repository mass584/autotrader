package service

import (
	"fmt"
	"time"

	"github.com/mass584/autotrader/entity"
	"github.com/mass584/autotrader/repository/database"
	"github.com/mass584/autotrader/repository/external/bitflyer"
	"github.com/mass584/autotrader/repository/external/coincheck"
	"github.com/mass584/autotrader/trade_algorithms"
	"gorm.io/gorm"
)

func DetermineOrderPrice() {
	orderBookBitflyer := bitflyer.GetOrderBook(entity.BTC_TO_JPY)
	tradesBitflyer := bitflyer.GetRecentTrades(entity.BTC_TO_JPY)
	orderPriceBitflyer := trade_algorithms.DetermineOrderPrice(orderBookBitflyer, tradesBitflyer.RecentTrades(5*time.Minute))
	fmt.Printf("Determined Order Price at Bitflyer: %.2f [JPY/BTC]\n", orderPriceBitflyer)

	orderBookCoinCheck := coincheck.GetOrderBook(entity.BTC_TO_JPY)
	tradesCoinCheck := coincheck.GetRecentTrades(entity.BTC_TO_JPY)
	orderPriceCoinCheck := trade_algorithms.DetermineOrderPrice(orderBookCoinCheck, tradesCoinCheck.RecentTrades(5*time.Minute))
	fmt.Printf("Determined Order Price at Coincheck: %.2f [JPY/BTC]\n", orderPriceCoinCheck)
}

func CalculateTradeSignalOnCoincheck(db *gorm.DB, exchangePair entity.ExchangePair, signalAt time.Time) {
	from := signalAt.Add(-50*24*time.Hour - 1*time.Minute)
	to := signalAt
	tradeCollection := database.GetTradesByTimeRange(db, entity.Coincheck, exchangePair, from, to)

	trendSignal, _ := trade_algorithms.TrendFollowingSignal(tradeCollection)
	fmt.Printf("TrandFollowSignal is: %s\n", trendSignal)
	meanReversionSignal, _ := trade_algorithms.MeanReversionSignal(tradeCollection)
	fmt.Printf("MeanReversionSignal is: %s\n", meanReversionSignal)
}
