package helper

import (
	"sort"
	"time"

	"github.com/mass584/autotrader/entity"
	"github.com/mass584/autotrader/repository/database"
	"github.com/mass584/autotrader/service"
	"github.com/rs/zerolog/log"
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
	sort.Slice(tradeCollection, func(a, b int) bool {
		return tradeCollection[a].Time.After(tradeCollection[b].Time)
	})

	_, err := database.SaveTrades(db, tradeCollection)
	if err != nil {
		log.Error().Err(err).Send()
	}
}

func AggregateHelper(
	db *gorm.DB,
	exchangePlace entity.ExchangePlace,
	exchangePair entity.ExchangePair,
	aggregateFrom time.Time,
	aggregateTo time.Time,
) {
	err := service.Aggregation(
		db,
		exchangePlace,
		exchangePair,
		aggregateFrom,
		aggregateTo,
	)
	if err != nil {
		log.Error().Err(err).Send()
	}
}
