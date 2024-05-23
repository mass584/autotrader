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
	ID             int            `gorm:"column:id"`
	ScrapingStatus ScrapingStatus `gorm:"column:scraping_status"`
	ExchangePlace  ExchangePlace  `gorm:"column:exchange_place"`
	ExchangePair   ExchangePair   `gorm:"column:exchange_pair"`
	FromID         int            `gorm:"column:from_id"`
	ToID           int            `gorm:"column:to_id"`
	FromTime       time.Time      `gorm:"column:from_time"`
	ToTime         time.Time      `gorm:"column:to_time"`
}
