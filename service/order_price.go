package service

import (
	"math"
	"time"

	"github.com/mass584/autotrader/entity"
	"github.com/mass584/autotrader/repository/external/coincheck"
	"github.com/rs/zerolog/log"
)

// このメソッドをよんでいるところはまだないが、実際の自動トレードで指値注文を出す場合に使う
func DetermineOrderPriceOnCoincheck(exchangePair entity.ExchangePair) float64 {
	orderBook := coincheck.GetOrderBook(exchangePair)
	trades := coincheck.GetRecentTrades(exchangePair)
	orderPrice := orderPrice(orderBook, trades.RecentTrades(5*time.Minute))
	log.Info().Msgf("Determined Order Price at Coincheck is %.2f [JPY/BTC]", orderPrice)
	return orderPrice
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
