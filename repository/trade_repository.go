package repository

import (
	"github.com/mass584/autotrader/config"
	"github.com/mass584/autotrader/entity"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func SaveTrades(tradeCollection entity.TradeCollection) {
	config, err := config.NewConfig()
	if err != nil {
		return
	}
	db, err := gorm.Open(mysql.Open(config.DatabaseURL()))
	if err != nil {
		return
	}
	db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "exchange_name"}, {Name: "trade_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"price", "volume", "time"}),
	}).Create(&tradeCollection)
}
