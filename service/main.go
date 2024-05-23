package service

import (
	"fmt"
	"sort"
	"time"

	"github.com/mass584/autotrader/entity"
	"github.com/mass584/autotrader/repository/database"
	"github.com/mass584/autotrader/repository/external/bitflyer"
	"github.com/mass584/autotrader/repository/external/coincheck"
	"github.com/mass584/autotrader/trade_algorithms"
	"gorm.io/gorm"
)

func ScrapingTradesFromCoincheck(db *gorm.DB, exchangePair entity.ExchangePair) {
	// スクレイピング履歴の取得
	scrapingHistories := database.GetScrapingHistoriesByStatus(
		db,
		entity.Coincheck,
		exchangePair,
		entity.ScrapingStatusSuccess,
	)
	sort.Slice(scrapingHistories, func(a, b int) bool {
		// 先頭が最新の取引履歴になるように、FromIDが大きい順に並べる
		return scrapingHistories[a].FromID > scrapingHistories[b].FromID
	})

	// スクレイピング範囲の決定
	var fromID int
	if len(scrapingHistories) > 0 {
		// 最新の取得履歴の次のIDから取得する
		// もし取得に失敗した範囲がある場合はその範囲も取得するべきだが、そのような処理はまだ入っていない
		fromID = scrapingHistories[0].ToID + 1
	} else {
		// 初回実行の時にはid=240000001(2023-02-22 19:03:39)まで遡る
		// 取引ペアが違くてもuniqueなIDが割り当てられているため、取引ペアによらずこのIDから取得する
		fromID = 240000001
	}
	toID := fromID + 1000000 - 1

	// スクレイピング履歴の作成
	tradeFrom := coincheck.GetAllTradesByLastId(exchangePair, fromID).LatestTrade()
	tradeTo := coincheck.GetAllTradesByLastId(exchangePair, toID).LatestTrade()
	scrapingHistory := entity.ScrapingHistory{
		ExchangePlace: entity.Coincheck,
		ExchangePair:  exchangePair,
		FromID:        tradeFrom.TradeID,
		ToID:          tradeTo.TradeID,
		FromTime:      tradeFrom.Time,
		ToTime:        tradeTo.Time,
	}
	history := database.CreateScrapingHistory(db, scrapingHistory)

	// スクレイピングの実行
	count := toID - fromID + 1
	per := 50 // スクレイピング対象となるAPIの都合上、50件ずつに分割して取得する
	pageMax := (count+1)/per + 1
	for page := 0; page < pageMax; page++ {
		// 実際はIDは通しではなく間が抜けて返ってくるので、同じものを重複取得してしまっており効率が悪くなっている
		lastId := fromID + page*per + per - 1
		time.Sleep(100 * time.Millisecond) // レートリミットに引っかからないように100ミリ秒待つ
		tradeCollection := coincheck.GetAllTradesByLastId(exchangePair, lastId)
		database.SaveTrades(db, tradeCollection)
	}

	// スクレイピングステータスの更新
	database.UpdateScrapingHistoryStatus(db, history, entity.ScrapingStatusSuccess)
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
