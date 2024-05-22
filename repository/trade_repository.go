package repository

import (
	"github.com/mass584/autotrader/entity"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func SaveTrades(db *gorm.DB, tradeCollection entity.TradeCollection) {
	db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "exchange_name"}, {Name: "trade_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"price", "volume", "time"}),
	}).Create(&tradeCollection)
}
