package service

import (
	"database/sql"
	"time"

	"github.com/mass584/autotrader/entity"
	"github.com/mass584/autotrader/repository/database"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

const FUND_MAX_YEN = 500000
const UNIT_VOLUME_YEN = 100000
const TAKE_PROFIT_AMOUNT_YEN = 20000
const STOP_LOSS_AMOUNT_YEN = 10000

func closePositions(
	db *gorm.DB,
	exchangePlace entity.ExchangePlace,
	exchangePair entity.ExchangePair,
	time time.Time,
) error {
	// 現在のポジションを取得
	positions, err := database.GetPositionsByStatus(
		db,
		exchangePlace,
		exchangePair,
		entity.PositionTypeLong,
		entity.PositionStatusHold,
	)
	if err != nil {
		return err
	}

	// 現在の価格を取得
	// 取引モデルのパラメータチューニングの際は、過去の指定日時の取引価格を取得するため、データベースから価格をひいている。
	// その際、正しく取得するためにはスクレイピング済みである必要があることに注意。
	// また、実際の取引の場合はWebSocketAPIなどでリアルタイムな価格を取得する必要があることに注意。
	trade, err := database.GetTradeByLatestBefore(db, exchangePlace, exchangePair, time)
	if err != nil {
		// 10分間取引がない場合は取得できなく、エラーとなる
		return err
	}

	currentPrice := trade.Price

	// 現在のポジションがクローズ対象かどうが判定して、そうであればクローズする
	// 一旦はロングポジションだけを考える
	failed := false
	for _, position := range positions {
		if currentPrice > position.BuyPrice.Float64 {
			// 利益確定条件を満たす場合はポジションをクローズする
			profit := currentPrice*position.Volume - position.BuyPrice.Float64*position.Volume
			if profit > TAKE_PROFIT_AMOUNT_YEN {
				// TODO 利益確定の注文リクエストを送信する処理をかく
				// 実際の取引の場合は、ここでスリッページが発生する可能性があることに注意
				position.PositionStatus = entity.PositionStatusClosedByTakeProfit
				position.SellPrice = sql.NullFloat64{Float64: currentPrice, Valid: true}
				position.SellTime = sql.NullTime{Time: time, Valid: true}
				_, err := database.SavePosition(db, position)
				if err != nil {
					failed = true
					log.Warn().Stack().Err(err).Send()
					continue
				}
			}
		} else if currentPrice < position.BuyPrice.Float64 {
			loss := position.BuyPrice.Float64*position.Volume - currentPrice*position.Volume
			// 損切り条件を満たす場合はポジションをクローズする
			if loss > STOP_LOSS_AMOUNT_YEN {
				// TODO 損切りの注文リクエストを送信する処理をかく
				// 実際の取引の場合は、ここでスリッページが発生する可能性があることに注意
				position.PositionStatus = entity.PositionStatusClosedByStopLoss
				position.SellPrice = sql.NullFloat64{Float64: currentPrice, Valid: true}
				position.SellTime = sql.NullTime{Time: time, Valid: true}
				_, err := database.SavePosition(db, position)
				if err != nil {
					failed = true
					log.Warn().Stack().Err(err).Send()
					continue
				}
			}
		}
	}

	if failed {
		err = errors.New("Failed to save position data.")
		return errors.WithStack(err)
	}

	return nil
}

func openPosition(
	db *gorm.DB,
	exchangePlace entity.ExchangePlace,
	exchangePair entity.ExchangePair,
	time time.Time,
) error {
	// 現在のポジションを取得
	positions, err := database.GetPositionsByStatus(
		db,
		exchangePlace,
		exchangePair,
		entity.PositionTypeLong,
		entity.PositionStatusHold,
	)
	if err != nil {
		return err
	}

	// 現在の価格を取得
	// 取引モデルのパラメータチューニングの際は、過去の指定日時の取引価格を取得するため、データベースから価格をひいている。
	// その際、正しく取得するためにはスクレイピング済みである必要があることに注意。
	// また、実際の取引の場合はWebSocketAPIなどでリアルタイムな価格を取得する必要があることに注意。
	trade, err := database.GetTradeByLatestBefore(db, exchangePlace, exchangePair, time)
	if err != nil {
		return err
	}

	currentPrice := trade.Price

	// ポジションが資金の上限を超える場合はここで終了
	var positionSum float64
	for _, position := range positions {
		positionSum += position.Volume * position.BuyPrice.Float64
	}

	tradeMargin := FUND_MAX_YEN - positionSum
	if UNIT_VOLUME_YEN > tradeMargin {
		return nil
	}

	// 新しいポジションを取得するかどうか判定して、そうであればリクエストする
	trendFollowSignal, err := trendFollowingSignal(db, exchangePlace, exchangePair, time)
	if err != nil {
		log.Warn().Stack().Err(err).Send()
	}

	// 一旦はトレンドフォローシグナルだけを見て新しいポジションを取得するかどうか判定しているが、
	// 実際には複数のシグナルを組み合わせて判定することが望ましい
	// 一旦はロングポジションだけを考える
	if trendFollowSignal == Buy {
		// TODO ロングポジションの買い注文リクエストを送信する処理をかく
		// 実際の取引の場合は、ここでスリッページが発生する可能性があることに注意
		newPosition := entity.Position{
			PositionType:   entity.PositionTypeLong,
			PositionStatus: entity.PositionStatusHold,
			ExchangePlace:  exchangePlace,
			ExchangePair:   exchangePair,
			// 一旦は現在価格で注文しているが、実際には板情報を使って指値注文を出すべき
			Volume:   UNIT_VOLUME_YEN / currentPrice,
			BuyPrice: sql.NullFloat64{Float64: currentPrice, Valid: true},
			BuyTime:  sql.NullTime{Time: time, Valid: true},
		}
		_, err := database.SavePosition(db, newPosition)

		if err != nil {
			return err
		}
	}

	return nil
}

func WatchPostion(db *gorm.DB, exchangePlace entity.ExchangePlace, exchangePair entity.ExchangePair) {
	for {
		at := time.Now()
		err := closePositions(db, exchangePlace, exchangePair, at)
		if err != nil {
			log.Warn().Stack().Err(err).Send()
		}

		err = openPosition(db, exchangePlace, exchangePair, at)
		if err != nil {
			log.Warn().Stack().Err(err).Send()
		}

		time.Sleep(1 * time.Minute)
	}
}

func simulationRange(exchangePlace entity.ExchangePlace) (time.Time, time.Time) {
	switch exchangePlace {
	case entity.Bitflyer:
		return time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC), time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	case entity.Coincheck:
		return time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC), time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC)
	default:
		return time.Now().UTC(), time.Now().UTC()
	}
}

func WatchPostionSimulation(db *gorm.DB, exchangePlace entity.ExchangePlace, exchangePair entity.ExchangePair) {
	simulationTime, simulationEnd := simulationRange(exchangePlace)
	for simulationTime.Before(simulationEnd) {
		simulationTime = simulationTime.Add(1 * time.Hour)
		err := closePositions(db, exchangePlace, exchangePair, simulationTime)
		if err != nil {
			log.Warn().Stack().Err(err).Send()
		}

		err = openPosition(db, exchangePlace, exchangePair, simulationTime)
		if err != nil {
			log.Warn().Stack().Err(err).Send()
		}
	}
}
