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
	// 過去に遡った取引履歴をスクレイピングするが
	// id=240000001(2023-02-22 19:03:39)よりも過去には遡らない
	fromID := 240000001
	// TODO fromIDをスクレイピング履歴から決定するように変更する
	// fromID := hogehoge

	// この関数の実行時間を短くするために、過去100000件の取引履歴をスクレイピングする
	// およそ3時間くらいかかる見込みだが、必要に応じて調整すること
	toID := fromID + 100000 - 1

	// スクレイピング履歴の作成
	tradeFrom := coincheck.GetAllTradesByLastId(entity.BTC_TO_JPY, fromID).LatestTrade()
	tradeTo := coincheck.GetAllTradesByLastId(entity.BTC_TO_JPY, toID).LatestTrade()
	scrapingHistory := entity.ScrapingHistory{
		FromID:   tradeFrom.ID,
		ToID:     tradeTo.ID,
		FromTime: tradeFrom.Time,
		ToTime:   tradeTo.Time,
	}
	database.CreateScrapingHistory(db, scrapingHistory)

	// スクレイピングの実行
	count := toID - fromID + 1
	per := 50 // スクレイピング対象となるAPIの都合上、50件ずつに分割して取得する
	pageMax := (count+1)/per + 1
	for page := 0; page < pageMax; page++ {
		lastId := fromID + page*per + per - 1
		time.Sleep(100 * time.Millisecond) // レートリミットに引っかからないように100ミリ秒待つ
		tradeCollection := coincheck.GetAllTradesByLastId(entity.BTC_TO_JPY, lastId)
		database.SaveTrades(db, tradeCollection)
	}

	// スクレイピングステータスの更新
	database.UpdateScrapingHistoryStatus(db, scrapingHistory, entity.ScrapingStatusSuccess)
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
