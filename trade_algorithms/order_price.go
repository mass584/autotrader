package trade_algorithms

import (
	"math"

	"github.com/mass584/auto-trade/entity"
)

func DetermineOrderPrice(orderBook entity.OrderBook, trades entity.TradeCollection) float64 {
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
