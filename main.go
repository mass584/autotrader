package main

import (
	"fmt"
	"time"

	"github.com/mass584/autotrader/entity"
	"github.com/mass584/autotrader/exchange/bitflyer"
	"github.com/mass584/autotrader/exchange/coincheck"
	"github.com/mass584/autotrader/repository"
	"github.com/mass584/autotrader/trade_algorithms"
)

func ScrapingFromCoinCheck() {
	startID := 1       // TODO データベースの値から決めるようにする
	endID := 264330000 // TODO APIで取得した値から決めるようにする
	count := endID - startID + 1
	per := 50
	pageMax := (count+1)/per + 1

	for page := 0; page < pageMax; page++ {
		lastId := startID + page*per + per - 1
		tradeCollection := coincheck.GetAllTradesByLastId(entity.BTC_TO_JPY, lastId)
		repository.SaveTrades(tradeCollection)
	}
}

func main() {
	orderBookBitflyer := bitflyer.GetOrderBook(entity.BTC_TO_JPY)
	tradesBitflyer := bitflyer.GetRecentTrades(entity.BTC_TO_JPY)
	orderPriceBitflyer := trade_algorithms.DetermineOrderPrice(orderBookBitflyer, tradesBitflyer.RecentTrades(5*time.Minute))
	fmt.Printf("Determined Order Price at Bitflyer: %.2f [JPY/BTC]\n", orderPriceBitflyer)

	orderBookCoinCheck := coincheck.GetOrderBook(entity.BTC_TO_JPY)
	tradesCoinCheck := coincheck.GetRecentTrades(entity.BTC_TO_JPY)
	orderPriceCoinCheck := trade_algorithms.DetermineOrderPrice(orderBookCoinCheck, tradesCoinCheck.RecentTrades(5*time.Minute))
	fmt.Printf("Determined Order Price at Coincheck: %.2f [JPY/BTC]\n", orderPriceCoinCheck)

	ScrapingFromCoinCheck()
}
