package entity

import "time"

type ScrapingStatus int

// DBに永続化されるので順番を変えないこと
const (
	ScrapingStatusProcessing ScrapingStatus = iota
	ScrapingStatusSuccess
	ScrapingStatusFailed
)

type ScrapingHistory struct {
	ID             int
	ScrapingStatus ScrapingStatus
	ExchangePlace  ExchangePlace
	ExchangePair   ExchangePair
	FromID         int
	ToID           int
	FromTime       time.Time
	ToTime         time.Time
}
