package service

import (
	"time"

	"github.com/mass584/autotrader/entity"
	"github.com/mass584/autotrader/repository/database"
	"gorm.io/gorm"
)

func Aggregation(
	db *gorm.DB,
	exchangePlace entity.ExchangePlace,
	exchangePair entity.ExchangePair,
	aggregateFrom time.Time,
	aggregateTo time.Time,
) error {
	startDate := aggregateFrom
	for {
		if startDate.After(aggregateTo) {
			break
		}
		newTradeAggregation, error := database.GenerateNewAggregation(db, exchangePlace, exchangePair, startDate)
		if error != nil {
			return error
		}

		_, error = database.SaveTradeAggregation(db, *newTradeAggregation)
		if error != nil {
			return error
		}

		startDate = startDate.Add(24 * time.Hour)
	}

	return nil
}

func aggregateFrom(exchangePlace entity.ExchangePlace) time.Time {
	switch exchangePlace {
	case entity.Bitflyer:
		return time.Date(2024, 4, 30, 0, 0, 0, 0, time.UTC)
	case entity.Coincheck:
		return time.Date(2023, 2, 23, 0, 0, 0, 0, time.UTC)
	default:
		return time.Now().UTC()
	}
}

func AggregationAll(db *gorm.DB, exchangePlace entity.ExchangePlace, exchangePair entity.ExchangePair) error {
	tradeAggregations, err := database.GetAllTradeAggregations(db, exchangePlace, exchangePair)
	if err != nil {
		return err
	}

	var from time.Time
	if len(tradeAggregations) == 0 {
		from = aggregateFrom(exchangePlace)
	} else {
		from = tradeAggregations[0].AggregateDate.Add(24 * time.Hour)
	}

	yesterday := time.Now().UTC().Add(-24 * time.Hour)
	year, month, day := yesterday.Date()
	to := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)

	return Aggregation(db, exchangePlace, exchangePair, from, to)
}
