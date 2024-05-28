package entity

import "time"

type TradeAggregation struct {
	ID               int
	ExchangePlace    ExchangePlace
	ExchangePair     ExchangePair
	AggregateDate    time.Time
	AveragePrice     float64
	TotalCount       int
	TotalTransaction float64
}
