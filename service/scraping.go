package service

import (
	"sort"
	"time"

	"github.com/mass584/autotrader/entity"
	"github.com/mass584/autotrader/repository/database"
	"github.com/mass584/autotrader/repository/external/coincheck"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

func ScrapingTradesFromCoincheck(db *gorm.DB, exchangePair entity.ExchangePair) error {
	// スクレイピング履歴の取得
	scrapingHistories, err := database.GetScrapingHistoriesByStatus(
		db,
		entity.Coincheck,
		exchangePair,
		entity.ScrapingStatusSuccess,
	)
	if err != nil {
		return errors.WithStack(err)
	}

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
	toID := fromID + 100000 - 1

	// スクレイピング履歴の作成
	var tradeCollection entity.TradeCollection
	tradeCollection, err = coincheck.GetAllTradesByLastId(exchangePair, fromID)
	if err != nil {
		return errors.WithStack(err)
	}
	tradeFrom := tradeCollection.LatestTrade()

	tradeCollection, err = coincheck.GetAllTradesByLastId(exchangePair, toID)
	if err != nil {
		return errors.WithStack(err)
	}
	tradeTo := tradeCollection.LatestTrade()

	scrapingHistory, err := database.SaveScrapingHistory(
		db,
		entity.ScrapingHistory{
			ExchangePlace: entity.Coincheck,
			ExchangePair:  exchangePair,
			FromID:        tradeFrom.TradeID,
			ToID:          tradeTo.TradeID,
			FromTime:      tradeFrom.Time,
			ToTime:        tradeTo.Time,
		})
	if err != nil {
		return errors.WithStack(err)
	}

	// スクレイピングの実行
	failed := false
	lastID := toID
	for lastID >= fromID {
		time.Sleep(100 * time.Millisecond) // レートリミットに引っかからないように100ミリ秒待つ

		tradeCollection, err := coincheck.GetAllTradesByLastId(exchangePair, lastID)
		if err != nil {
			failed = true
			log.Warn().Err(err).Msgf("Failed to get trades from Coincheck. lastID=%d", lastID)
			continue
		}

		_, err = database.SaveTrades(db, tradeCollection)
		if err != nil {
			failed = true
			log.Warn().Err(err).Msgf("Failed to save trades. lastID=%d", lastID)
			continue
		}

		lastID = tradeCollection.OldestTrade().TradeID
	}

	// スクレイピングステータスの更新
	if failed {
		scrapingHistory.ScrapingStatus = entity.ScrapingStatusFailed
	} else {
		scrapingHistory.ScrapingStatus = entity.ScrapingStatusSuccess
	}
	_, err = database.SaveScrapingHistory(db, *scrapingHistory)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
