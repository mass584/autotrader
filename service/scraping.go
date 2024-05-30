package service

import (
	"errors"
	"sort"
	"time"

	"github.com/mass584/autotrader/entity"
	"github.com/mass584/autotrader/repository/database"
	"github.com/mass584/autotrader/repository/external/coincheck"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

var ErrUnsupportedExchangePlace = errors.New("unsupported exchange place")

func getScrapingRangeFromCoincheck(
	exchangePair entity.ExchangePair,
	scrapingHistories []entity.ScrapingHistory,
) (*entity.Trade, *entity.Trade, error) {
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

	var tradeCollection entity.TradeCollection
	tradeCollection, err := coincheck.GetAllTradesByLastId(exchangePair, fromID)
	if err != nil {
		return nil, nil, err
	}
	tradeFrom := tradeCollection.LatestTrade()

	tradeCollection, err = coincheck.GetAllTradesByLastId(exchangePair, toID)
	if err != nil {
		return nil, nil, err
	}
	tradeTo := tradeCollection.LatestTrade()

	return &tradeFrom, &tradeTo, nil
}

func execScrapingFromCoincheck(db *gorm.DB, exchangePair entity.ExchangePair, tradeFrom, tradeTo *entity.Trade) bool {
	dirty := false
	lastID := tradeTo.TradeID
	for lastID >= tradeFrom.TradeID {
		time.Sleep(100 * time.Millisecond) // レートリミットに引っかからないように100ミリ秒待つ

		tradeCollection, err := coincheck.GetAllTradesByLastId(exchangePair, lastID)
		if err != nil {
			dirty = true
			log.Warn().Err(err).Msgf("Failed to get trades from Coincheck. lastID=%d", lastID)
			continue // 失敗しても中断しないで続行する
		}

		_, err = database.SaveTrades(db, tradeCollection)
		if err != nil {
			dirty = true
			log.Warn().Err(err).Msgf("Failed to save trades. lastID=%d", lastID)
			continue // 失敗しても中断しないで続行する
		}

		lastID = tradeCollection.OldestTrade().TradeID
	}

	return dirty
}

func ScrapingTrades(
	db *gorm.DB,
	exchangePlace entity.ExchangePlace,
	exchangePair entity.ExchangePair,
) error {
	// スクレイピング履歴の取得
	scrapingHistories, err := database.GetScrapingHistoriesByStatus(
		db,
		exchangePlace,
		exchangePair,
		entity.ScrapingStatusSuccess,
	)
	if err != nil {
		return err
	}

	sort.Slice(scrapingHistories, func(a, b int) bool {
		// 先頭が最新の取引履歴になるように、FromIDが大きい順に並べる
		return scrapingHistories[a].FromID > scrapingHistories[b].FromID
	})

	var tradeFrom, tradeTo *entity.Trade
	switch exchangePlace {
	case entity.Coincheck:
		tradeFrom, tradeTo, err = getScrapingRangeFromCoincheck(exchangePair, scrapingHistories)
	default:
		return ErrUnsupportedExchangePlace
	}
	if err != nil {
		return err
	}

	// スクレイピング履歴の作成
	scrapingHistory, err := database.SaveScrapingHistory(
		db,
		entity.ScrapingHistory{
			ExchangePlace: exchangePlace,
			ExchangePair:  exchangePair,
			FromID:        tradeFrom.TradeID,
			ToID:          tradeTo.TradeID,
			FromTime:      tradeFrom.Time,
			ToTime:        tradeTo.Time,
		})
	if err != nil {
		return err
	}

	// スクレイピングの実行
	var dirty bool

	switch exchangePlace {
	case entity.Coincheck:
		dirty = execScrapingFromCoincheck(db, exchangePair, tradeFrom, tradeTo)
	}

	// スクレイピングステータスの更新
	if dirty {
		scrapingHistory.ScrapingStatus = entity.ScrapingStatusFailed
	} else {
		scrapingHistory.ScrapingStatus = entity.ScrapingStatusSuccess
	}
	_, err = database.SaveScrapingHistory(db, *scrapingHistory)
	if err != nil {
		return err
	}

	return nil
}
