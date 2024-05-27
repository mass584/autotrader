package entity

import "time"

type PositionType int
type PositionStatus int

// DBに永続化されるので順番を変えないこと
const (
	PositionTypeLong PositionType = iota
	PositionTypeShort
)

// DBに永続化されるので順番を変えないこと
const (
	PositionStatusHold PositionStatus = iota
	PositionStatusClosedByTakeProfit
	PositionStatusClosedByStopLoss
)

type Position struct {
	ID             int
	PositionType   PositionType
	PositionStatus PositionStatus
	ExchangePlace  ExchangePlace
	ExchangePair   ExchangePair
	Volume         float64
	BuyPrice       float64
	SellPrice      float64
	BuyTime        time.Time
	SellTime       time.Time
}
