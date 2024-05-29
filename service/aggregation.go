package service

import (
	"time"

	"github.com/mass584/autotrader/entity"
	"github.com/mass584/autotrader/repository/database"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

func Aggregation(
	db *gorm.DB,
	exchangePlace entity.ExchangePlace,
	exchangePair entity.ExchangePair,
	aggregateFrom time.Time,
	aggregateTo time.Time,
) {
	startDate := aggregateFrom
	for {
		if startDate.After(aggregateTo) {
			log.Info().Msg("Aggregation is completed.")
			break
		}
		newTradeAggregation, error := database.GenerateNewAggregation(db, exchangePlace, exchangePair, startDate)
		if error != nil {
			log.Info().Msgf("%v", error)
			break
		}
		database.SaveTradeAggregation(db, *newTradeAggregation)
		startDate = startDate.Add(24 * time.Hour)
	}
}

func AggregationAll(
	db *gorm.DB,
	exchangePlace entity.ExchangePlace,
	exchangePair entity.ExchangePair,
) {
	var from time.Time
	tradeAggregations := database.GetAllTradeAggregations(db, entity.Coincheck, exchangePair)
	if len(tradeAggregations) == 0 {
		from = time.Date(2023, 2, 23, 0, 0, 0, 0, time.UTC)
	} else {
		from = tradeAggregations[0].AggregateDate.Add(24 * time.Hour)
	}

	yesterday := time.Now().UTC().Add(-24 * time.Hour)
	year, month, day := yesterday.Date()
	to := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)

	Aggregation(db, exchangePlace, exchangePair, from, to)
}
