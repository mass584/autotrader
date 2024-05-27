package service

import (
	"fmt"
	"math"
	"time"

	"github.com/mass584/autotrader/entity"
	"github.com/mass584/autotrader/repository/external/bitflyer"
	"github.com/mass584/autotrader/repository/external/coincheck"
)

func DetermineOrderPrice() {
	orderBookBitflyer := bitflyer.GetOrderBook(entity.BTC_TO_JPY)
	tradesBitflyer := bitflyer.GetRecentTrades(entity.BTC_TO_JPY)
	orderPriceBitflyer := orderPrice(orderBookBitflyer, tradesBitflyer.RecentTrades(5*time.Minute))
	fmt.Printf("Determined Order Price at Bitflyer: %.2f [JPY/BTC]\n", orderPriceBitflyer)

	orderBookCoinCheck := coincheck.GetOrderBook(entity.BTC_TO_JPY)
	tradesCoinCheck := coincheck.GetRecentTrades(entity.BTC_TO_JPY)
	orderPriceCoinCheck := orderPrice(orderBookCoinCheck, tradesCoinCheck.RecentTrades(5*time.Minute))
	fmt.Printf("Determined Order Price at Coincheck: %.2f [JPY/BTC]\n", orderPriceCoinCheck)
}

func orderPrice(orderBook entity.OrderBook, trades entity.TradeCollection) float64 {
	bestBid := orderBook.Bids[0].Price
	bestAsk := orderBook.Asks[0].Price

	// スプレッドを縮小する
	orderPrice := (bestBid + bestAsk) / 2.0

	// 最近の取引価格を考慮
	if len(trades) > 0 {
		avgRecentPrice := 0.0
		for _, trade := range trades {
			avgRecentPrice += trade.Price
		}
		avgRecentPrice /= float64(len(trades))

		orderPrice = (orderPrice + avgRecentPrice) / 2.0
	}

	return math.Round(orderPrice*100) / 100 // 小数点以下2桁に丸める
}
