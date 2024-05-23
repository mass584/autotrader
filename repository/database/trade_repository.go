package database

import (
	"time"

	"github.com/mass584/autotrader/entity"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func SaveTrades(db *gorm.DB, tradeCollection entity.TradeCollection) {
	db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "exchange_place"}, {Name: "exchange_pair"}, {Name: "trade_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"price", "volume", "time"}),
	}).Create(&tradeCollection)
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
		Find(&tradeCollection)
	return tradeCollection
}
