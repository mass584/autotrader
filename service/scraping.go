package service

import (
	"errors"
	"time"

	"github.com/mass584/autotrader/entity"
	"github.com/mass584/autotrader/repository/database"
	"github.com/mass584/autotrader/repository/external/bitflyer"
	"github.com/mass584/autotrader/repository/external/coincheck"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

var ErrUnsupportedExchangePlace = errors.New("unsupported exchange place")
var ErrPendingScraping = errors.New("pending scraping")

// 取引所に依存した処理を実装する際のインターフェース
type ExchangePlaceFunctions interface {
	// 新しいスクレイピング履歴を生成する関数
	generateNewScrapingHistory(exchangePair entity.ExchangePair, scrapingHistories []entity.ScrapingHistory) (*entity.ScrapingHistory, error)
	// スクレイピングを実行する関数、戻り値はスクレイピングに失敗したかどうか
	execScraping(db *gorm.DB, exchangePair entity.ExchangePair, fromID, toID int) bool
}

func NewExchangePlaceFunctions(exchangePlace entity.ExchangePlace) ExchangePlaceFunctions {
	// 新しい取引所に対応する際はここに追加する
	switch exchangePlace {
	case entity.Bitflyer:
		return &BitflyerFunctions{}
	case entity.Coincheck:
		return &CoincheckFunctions{}
	default:
		return nil
	}
}

// 取引所ごとの処理を実装する
type BitflyerFunctions struct{}

func (_ *BitflyerFunctions) generateNewScrapingHistory(
	exchangePair entity.ExchangePair,
	scrapingHistories []entity.ScrapingHistory,
) (*entity.ScrapingHistory, error) {
	var fromID, toID int
	if len(scrapingHistories) > 0 {
		// 最新の取得履歴の次のIDから取得する
		fromID = scrapingHistories[0].ToID + 1
		toID = fromID + 100000 - 1
	} else {
		// 初回実行の時にはid=2522208992(2024-04-29 04:06:06)まで遡る
		// 取引ペアが違くてもuniqueなIDが割り当てられているため、取引ペアによらずこのIDから取得する
		// 最大31日までしか遡れないようになっている
		fromID = 2522208992
		toID = fromID + 100000 - 1
	}

	var tradeFrom, tradeTo entity.Trade
	for {
		time.Sleep(1000 * time.Millisecond) // レートリミットに引っかからないように1000ミリ秒待つ

		var tradeCollection entity.TradeCollection
		tradeCollection, err := bitflyer.GetTradesByLastID(exchangePair, fromID)
		if err == bitflyer.ErrIDIsTooOld {
			// スクレイピング範囲が31日よりも前の場合は取得できないので、スクレイピング範囲を進める
			toID += 100000
			fromID += 100000
			continue
		}
		if err != nil {
			return nil, err
		}
		tradeFrom = tradeCollection.LatestTrade()

		tradeCollection, err = bitflyer.GetTradesByLastID(exchangePair, toID)
		if err != nil {
			return nil, err
		}
		tradeTo = tradeCollection.LatestTrade()

		break
	}
	return &entity.ScrapingHistory{
		ExchangePlace: entity.Bitflyer,
		ExchangePair:  exchangePair,
		FromID:        fromID,
		ToID:          toID,
		FromTime:      tradeFrom.Time,
		ToTime:        tradeTo.Time,
	}, nil
}

func (_ *BitflyerFunctions) execScraping(db *gorm.DB, exchangePair entity.ExchangePair, fromID, toID int) bool {
	dirty := false
	lastID := toID
	for lastID >= fromID {
		time.Sleep(1000 * time.Millisecond) // レートリミットに引っかからないように1000ミリ秒待つ

		tradeCollection, err := bitflyer.GetTradesByLastID(exchangePair, lastID)
		if err != nil {
			dirty = true
			log.Warn().Err(err).Msgf("Failed to get trades from Bitflyer. lastID=%d", lastID)
			continue // 失敗しても中断しないで続行する
		}

		_, err = database.SaveTrades(db, tradeCollection)
		if err != nil {
			dirty = true
			log.Warn().Err(err).Msgf("Failed to save trades. lastID=%d", lastID)
			continue // 失敗しても中断しないで続行する
		}

		lastID = tradeCollection.OldestTrade().TradeID - 1
	}

	return dirty
}

type CoincheckFunctions struct{}

func (_ *CoincheckFunctions) generateNewScrapingHistory(
	exchangePair entity.ExchangePair,
	scrapingHistories []entity.ScrapingHistory,
) (*entity.ScrapingHistory, error) {
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
		return nil, err
	}
	tradeFrom := tradeCollection.LatestTrade()

	tradeCollection, err = coincheck.GetAllTradesByLastId(exchangePair, toID)
	if err != nil {
		return nil, err
	}
	tradeTo := tradeCollection.LatestTrade()

	return &entity.ScrapingHistory{
		ExchangePlace: entity.Coincheck,
		ExchangePair:  exchangePair,
		FromID:        fromID,
		ToID:          toID,
		FromTime:      tradeFrom.Time,
		ToTime:        tradeTo.Time,
	}, nil
}

func (_ *CoincheckFunctions) execScraping(db *gorm.DB, exchangePair entity.ExchangePair, fromID, toID int) bool {
	dirty := false
	lastID := toID
	for lastID >= fromID {
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

func scrapingOneBlock(
	db *gorm.DB,
	exchangePlace entity.ExchangePlace,
	exchangePair entity.ExchangePair,
) error {
	funcs := NewExchangePlaceFunctions(exchangePlace)

	scrapingHistories, err := database.GetScrapingHistoriesByStatus(
		db,
		exchangePlace,
		exchangePair,
		entity.ScrapingStatusSuccess,
	)
	if err != nil {
		return err
	}

	newScrapingHistory, err := funcs.generateNewScrapingHistory(exchangePair, scrapingHistories)
	if err != nil {
		return err
	}

	maxTime := time.Now().Add(-24 * time.Hour).UTC()
	if newScrapingHistory.FromTime.After(maxTime) {
		return ErrPendingScraping
	}

	scrapingHistory, err := database.SaveScrapingHistory(
		db,
		*newScrapingHistory,
	)
	if err != nil {
		return err
	}

	dirty := funcs.execScraping(db, exchangePair, scrapingHistory.FromID, scrapingHistory.ToID)
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

func ScrapingTrades(
	db *gorm.DB,
	exchangePlace entity.ExchangePlace,
	exchangePair entity.ExchangePair,
) {
	for {
		err := scrapingOneBlock(db, exchangePlace, exchangePair)
		if err != nil {
			log.Error().Stack().Err(err).Send()
		}

		if errors.Is(err, ErrPendingScraping) {
			time.Sleep(24 * time.Hour)
		}

		time.Sleep(10 * time.Second)
	}
}
