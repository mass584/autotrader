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

func ScrapingTradesFromCoincheck(db *gorm.DB) {
	startID := 240000000 // TODO データベースの値から決めるようにする
	endID := 264330000   // TODO APIで取得した値から決めるようにする
	count := endID - startID + 1
	per := 50
	pageMax := (count+1)/per + 1

	for page := 0; page < pageMax; page++ {
		lastId := startID + page*per + per - 1
		time.Sleep(100 * time.Millisecond)
		tradeCollection := coincheck.GetAllTradesByLastId(entity.BTC_TO_JPY, lastId)
		database.SaveTrades(db, tradeCollection)
	}
}

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

func CalculateTradeSignal() {
	// このリポジトリは取引所のサーバーからスクレイピングする実装になっているので、
	// 50件しか取得できないため、移動平均を計算するのに十分なコレクションを取得することができない。
	// データベースに永続化したストアから取得するリポジトリに差し替える必要がある。
	tradeCollection := coincheck.GetAllTradesByLastId(entity.BTC_TO_JPY, 264330000)
	trendSignal, _ := trade_algorithms.TrendFollowingSignal(tradeCollection)
	fmt.Printf("TrandFollowSignal is: %s\n", trendSignal)
	meanReversionSignal, _ := trade_algorithms.MeanReversionSignal(tradeCollection)
	fmt.Printf("MeanReversionSignal is: %s\n", meanReversionSignal)
}
