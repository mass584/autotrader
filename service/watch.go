package service

import (
	"database/sql"
	"time"

	"github.com/mass584/autotrader/entity"
	"github.com/mass584/autotrader/repository/database"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

const FUND_MAX_YEN = 500000
const UNIT_VOLUME_YEN = 100000
const TAKE_PROFIT_AMOUNT_YEN = 20000
const STOP_LOSS_AMOUNT_YEN = 10000

func closePositions(db *gorm.DB, time time.Time) {
	// 現在のポジションを取得
	positions := database.GetPositionsByStatus(
		db,
		entity.Coincheck,
		entity.BTC_JPY,
		entity.PositionTypeLong,
		entity.PositionStatusHold,
	)

	// 現在の価格を取得
	// 取引モデルのパラメータチューニングの際は、過去の指定日時の取引価格を取得するため、データベースから価格をひいている。
	// その際、正しく取得するためにはスクレイピング済みである必要があることに注意。
	// また、実際の取引の場合はWebSocketAPIなどでリアルタイムな価格を取得する必要があることに注意。
	trade, error := database.GetTradeByLatestBefore(db, entity.Coincheck, entity.BTC_JPY, time)
	if error != nil {
		// 10分間取引がない場合は取得できなく、エラーとなる
		log.Warn().Msg("Failed to get trade data.")
		return
	}

	currentPrice := trade.Price

	// 現在のポジションがクローズ対象かどうが判定して、そうであればクローズする
	// 一旦はロングポジションだけを考える
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
				database.SavePosition(db, position)
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
				database.SavePosition(db, position)
			}
		}
	}

}

func openPosition(db *gorm.DB, time time.Time) {
	// 現在のポジションを取得
	positions := database.GetPositionsByStatus(
		db,
		entity.Coincheck,
		entity.BTC_JPY,
		entity.PositionTypeLong,
		entity.PositionStatusHold,
	)

	// 現在の価格を取得
	// 取引モデルのパラメータチューニングの際は、過去の指定日時の取引価格を取得するため、データベースから価格をひいている。
	// その際、正しく取得するためにはスクレイピング済みである必要があることに注意。
	// また、実際の取引の場合はWebSocketAPIなどでリアルタイムな価格を取得する必要があることに注意。
	trade, error := database.GetTradeByLatestBefore(db, entity.Coincheck, entity.BTC_JPY, time)
	if error != nil {
		log.Error().Msg("Failed to get trade data.")
		return
	}

	currentPrice := trade.Price

	// ポジションが資金の上限を超える場合はここで終了
	var positionSum float64
	for _, position := range positions {
		positionSum += position.Volume * position.BuyPrice.Float64
	}

	tradeMargin := FUND_MAX_YEN - positionSum
	if UNIT_VOLUME_YEN > tradeMargin {
		return
	}

	// 新しいポジションを取得するかどうか判定して、そうであればリクエストする
	trendFollowSignal := trendFollowingSignal(db, entity.Coincheck, entity.BTC_JPY, time)
	// 一旦はトレンドフォローシグナルだけを見て新しいポジションを取得するかどうか判定しているが、
	// 実際には複数のシグナルを組み合わせて判定することが望ましい
	// 一旦はロングポジションだけを考える
	if trendFollowSignal == Buy {
		// TODO ロングポジションの買い注文リクエストを送信する処理をかく
		// 実際の取引の場合は、ここでスリッページが発生する可能性があることに注意
		newPosition := entity.Position{
			PositionType:   entity.PositionTypeLong,
			PositionStatus: entity.PositionStatusHold,
			ExchangePlace:  entity.Coincheck,
			ExchangePair:   entity.BTC_JPY,
			// 一旦は現在価格で注文しているが、実際には板情報を使って指値注文を出すべき
			Volume:   UNIT_VOLUME_YEN / currentPrice,
			BuyPrice: sql.NullFloat64{Float64: currentPrice, Valid: true},
			BuyTime:  sql.NullTime{Time: time, Valid: true},
		}
		database.SavePosition(db, newPosition)
	}

}

func WatchPostionOnCoincheck(db *gorm.DB) {
	for {
		at := time.Now()
		closePositions(db, at)
		openPosition(db, at)
		time.Sleep(1 * time.Minute)
	}
}

func WatchPostionOnCoincheckForSimulation(db *gorm.DB) {
	at := time.Date(2023, 10, 1, 10, 0, 0, 0, time.Local)
	for {
		at = at.Add(1 * time.Hour)
		closePositions(db, at)
		openPosition(db, at)
	}
}
