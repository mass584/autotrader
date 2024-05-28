package database

import (
	"time"

	"github.com/mass584/autotrader/entity"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func GenerateNewAggregation(
	db *gorm.DB,
	exchangePlace entity.ExchangePlace,
	exchangePair entity.ExchangePair,
	date time.Time,
) (*entity.TradeAggregation, error) {
	var result struct {
		AveragePrice     float64
		TotalCount       int
		TotalTransaction float64
	}

	from := date
	to := date.Add(24 * time.Hour)
	db.
		Model(&entity.Trade{}).
		Where("exchange_place = ?", exchangePlace).
		Where("exchange_pair = ?", exchangePair).
		Where("? <= time and time < ?", from, to).
		Select("sum(price)/count(*) as average_price, count(*) as total_count, sum(price*volume) as total_transaction").
		Scan(&result)

	if result.TotalCount == 0 {
		return nil, errors.New("No trade data.")
	}

	return &entity.TradeAggregation{
		ExchangePlace:    exchangePlace,
		ExchangePair:     exchangePair,
		AggregateDate:    date,
		AveragePrice:     result.AveragePrice,
		TotalCount:       result.TotalCount,
		TotalTransaction: result.TotalTransaction,
	}, nil
}

func SaveTradeAggregation(db *gorm.DB, tradeAggregation entity.TradeAggregation) {
	db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "exchange_place"}, {Name: "exchange_pair"}, {Name: "aggregate_date"}},
		DoUpdates: clause.AssignmentColumns([]string{"average_price", "total_count", "total_transaction"}),
	}).Create(&tradeAggregation)
}

func GetAllTradeAggregations(
	db *gorm.DB,
	exchange_place entity.ExchangePlace,
	exchange_pair entity.ExchangePair,
) []entity.TradeAggregation {
	var tradeAggregations []entity.TradeAggregation
	db.
		Where("exchange_place = ?", exchange_place).
		Where("exchange_pair = ?", exchange_pair).
		Order("aggregate_date DESC").
		Find(&tradeAggregations)
	return tradeAggregations
}

func GetTradeAggregationsByTimeRange(
	db *gorm.DB,
	exchange_place entity.ExchangePlace,
	exchange_pair entity.ExchangePair,
	from time.Time,
	to time.Time,
) []entity.TradeAggregation {
	var tradeAggregations []entity.TradeAggregation
	db.
		Where("exchange_place = ?", exchange_place).
		Where("exchange_pair = ?", exchange_pair).
		Where("? <= time and time <= ?", from, to).
		Order("aggregate_date DESC").
		Find(&tradeAggregations)

	return tradeAggregations
}
