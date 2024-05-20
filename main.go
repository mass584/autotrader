package main

import (
	"fmt"
	"time"

	"github.com/mass584/autotrader/entity"
	"github.com/mass584/autotrader/exchange/bitflyer"
	"github.com/mass584/autotrader/exchange/coincheck"
	"github.com/mass584/autotrader/trade_algorithms"
)

func main() {
	orderBookBitflyer := bitflyer.GetOrderBook(entity.BTC_TO_JPY)
	tradesBitflyer := bitflyer.GetRecentTrades(entity.BTC_TO_JPY)
	orderPriceBitflyer := trade_algorithms.DetermineOrderPrice(orderBookBitflyer, tradesBitflyer.RecentTrades(5*time.Minute))
	fmt.Printf("Determined Order Price at Bitflyer: %.2f [JPY/BTC]\n", orderPriceBitflyer)

	orderBookCoinCheck := coincheck.GetOrderBook(entity.BTC_TO_JPY)
	tradesCoinCheck := coincheck.GetRecentTrades(entity.BTC_TO_JPY)
	orderPriceCoinCheck := trade_algorithms.DetermineOrderPrice(orderBookCoinCheck, tradesCoinCheck.RecentTrades(5*time.Minute))
	fmt.Printf("Determined Order Price at Coincheck: %.2f [JPY/BTC]\n", orderPriceCoinCheck)
}
