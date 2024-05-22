package main

import (
	"fmt"
	"time"

	"github.com/mass584/autotrader/config"
	"github.com/mass584/autotrader/entity"
	"github.com/mass584/autotrader/exchange/bitflyer"
	"github.com/mass584/autotrader/exchange/coincheck"
	"github.com/mass584/autotrader/repository"
	"github.com/mass584/autotrader/trade_algorithms"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func ScrapingFromCoinCheck(db *gorm.DB) {
	startID := 240000000 // TODO データベースの値から決めるようにする
	endID := 264330000   // TODO APIで取得した値から決めるようにする
	count := endID - startID + 1
	per := 50
	pageMax := (count+1)/per + 1

	for page := 0; page < pageMax; page++ {
		lastId := startID + page*per + per - 1
		time.Sleep(100 * time.Millisecond)
		tradeCollection := coincheck.GetAllTradesByLastId(entity.BTC_TO_JPY, lastId)
		repository.SaveTrades(db, tradeCollection)
	}
}

func main() {
	config, err := config.NewConfig()
	if err != nil {
		return
	}
	db, err := gorm.Open(mysql.Open(config.DatabaseURL()))
	if err != nil {
		return
	}

	ScrapingFromCoinCheck(db)

	orderBookBitflyer := bitflyer.GetOrderBook(entity.BTC_TO_JPY)
	tradesBitflyer := bitflyer.GetRecentTrades(entity.BTC_TO_JPY)
	orderPriceBitflyer := trade_algorithms.DetermineOrderPrice(orderBookBitflyer, tradesBitflyer.RecentTrades(5*time.Minute))
	fmt.Printf("Determined Order Price at Bitflyer: %.2f [JPY/BTC]\n", orderPriceBitflyer)

	orderBookCoinCheck := coincheck.GetOrderBook(entity.BTC_TO_JPY)
	tradesCoinCheck := coincheck.GetRecentTrades(entity.BTC_TO_JPY)
	orderPriceCoinCheck := trade_algorithms.DetermineOrderPrice(orderBookCoinCheck, tradesCoinCheck.RecentTrades(5*time.Minute))
	fmt.Printf("Determined Order Price at Coincheck: %.2f [JPY/BTC]\n", orderPriceCoinCheck)
}
