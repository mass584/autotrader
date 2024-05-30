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

func AggregationAllCoincheck(db *gorm.DB, exchangePair entity.ExchangePair) error {
	tradeAggregations, err := database.GetAllTradeAggregations(db, entity.Coincheck, exchangePair)
	if err != nil {
		return err
	}

	var from time.Time
	if len(tradeAggregations) == 0 {
		from = time.Date(2023, 2, 23, 0, 0, 0, 0, time.UTC)
	} else {
		from = tradeAggregations[0].AggregateDate.Add(24 * time.Hour)
	}

	yesterday := time.Now().UTC().Add(-24 * time.Hour)
	year, month, day := yesterday.Date()
	to := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)

	return Aggregation(db, entity.Coincheck, exchangePair, from, to)
}
