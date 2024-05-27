package entity

import "time"

type PositionStatus int

// DBに永続化されるので順番を変えないこと
const (
	PositionStatusProcessing PositionStatus = iota
	PositionStatusSuccess
	PositionStatusFailed
)

type Position struct {
	ID             int
	PositionStatus PositionStatus
	ExchangePlace  ExchangePlace
	ExchangePair   ExchangePair
	TradeID        int
	Price          float64
	Volume         float64
	Time           time.Time
}
