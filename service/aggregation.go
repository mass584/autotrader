package service

import (
	"time"

	"github.com/mass584/autotrader/entity"
	"github.com/mass584/autotrader/repository/database"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

func Aggregation(db *gorm.DB, exchangePlace entity.ExchangePlace, exchangePair entity.ExchangePair) {
	tradeAggregations := database.GetAllTradeAggregations(db, exchangePlace, exchangePair)

	var startDate time.Time
	if len(tradeAggregations) == 0 {
		startDate = time.Date(2023, 2, 23, 0, 0, 0, 0, time.UTC)
	} else {
		startDate = tradeAggregations[0].AggregateDate
	}

	now := time.Now().UTC()
	year, month, day := now.Date()
	today := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)

	for {
		if !startDate.Before(today) {
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
