package helper

import (
	"time"

	"github.com/mass584/autotrader/entity"
	"github.com/mass584/autotrader/repository/database"
	"github.com/mass584/autotrader/service"
	"gorm.io/gorm"
)

type Trades []struct {
	Price  float64
	Volume float64
	Time   time.Time
}

func BuildTradeCollectionHelper(trades Trades) entity.TradeCollection {
	baseTradeID := int(time.Now().UnixMilli())

	var tradeCollection entity.TradeCollection

	for idx, trade := range trades {
		tradeCollection = append(tradeCollection,
			entity.Trade{
				ExchangePlace: entity.Coincheck,
				ExchangePair:  entity.BTC_JPY,
				TradeID:       baseTradeID + idx,
				Price:         trade.Price,
				Volume:        trade.Volume,
				Time:          trade.Time,
			},
		)
	}

	return tradeCollection
}

func InsertTradeCollectionHelper(db *gorm.DB, tradeCollection entity.TradeCollection) {
	database.SaveTrades(db, tradeCollection)
}

func AggregateHelper(
	db *gorm.DB,
	exchangePlace entity.ExchangePlace,
	exchangePair entity.ExchangePair,
	aggregateFrom time.Time,
	aggregateTo time.Time,
) {
	service.Aggregation(
		db,
		exchangePlace,
		exchangePair,
		aggregateFrom,
		aggregateTo,
	)
}
