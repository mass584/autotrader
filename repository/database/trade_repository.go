package database

import (
	"time"

	"github.com/mass584/autotrader/entity"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func SaveTrades(db *gorm.DB, tradeCollection entity.TradeCollection) (entity.TradeCollection, error) {
	result := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "exchange_place"}, {Name: "exchange_pair"}, {Name: "trade_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"price", "volume", "time"}),
	}).Create(&tradeCollection)

	if result.Error != nil {
		return nil, errors.WithStack(result.Error)
	}

	return tradeCollection, nil
}

func GetTradesByTimeRange(
	db *gorm.DB,
	exchange_place entity.ExchangePlace,
	exchange_pair entity.ExchangePair,
	from time.Time,
	to time.Time,
) entity.TradeCollection {
	var tradeCollection entity.TradeCollection
	db.
		Where("exchange_place = ?", exchange_place).
		Where("exchange_pair = ?", exchange_pair).
		Where("? <= time and time <= ?", from, to).
		Order("time DESC").
		Find(&tradeCollection)
	return tradeCollection
}

func GetTradeByLatestBefore(
	db *gorm.DB,
	exchange_place entity.ExchangePlace,
	exchange_pair entity.ExchangePair,
	at time.Time,
) (*entity.Trade, error) {
	// ソートに時間がかかりすぎるので、10分前までのデータを取得する
	timeLeft := at.Add(-10 * time.Minute)

	var trade entity.Trade
	result := db.
		Where("exchange_place = ?", exchange_place).
		Where("exchange_pair = ?", exchange_pair).
		Where("? <= time and time <= ?", timeLeft, at).
		Order("time DESC").
		First(&trade)

	if result.Error != nil {
		return nil, errors.WithStack(result.Error)
	}

	return &trade, nil
}
