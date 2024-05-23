package database

import (
	"github.com/mass584/autotrader/entity"
	"gorm.io/gorm"
)

func CreateScrapingHistory(db *gorm.DB, scrapingHistory entity.ScrapingHistory) {
	scrapingHistory.ScrapingStatus = entity.ScrapingStatusProcessing
	db.Create(&scrapingHistory)
}

func UpdateScrapingHistoryStatus(db *gorm.DB, scrapingHistory entity.ScrapingHistory, status entity.ScrapingStatus) {
	scrapingHistory.ScrapingStatus = status
	db.Save(&scrapingHistory)
}

func GetScrapingHistoriesByStatus(
	db *gorm.DB,
	exchange_place entity.ExchangePlace,
	exchange_pair entity.ExchangePair,
	status entity.ScrapingStatus,
) []entity.ScrapingHistory {
	var scrapingHistories []entity.ScrapingHistory
	db.
		Where("exchange_place = ?", exchange_place).
		Where("exchange_pair = ?", exchange_pair).
		Where("scraping_status = ?", status).Find(&scrapingHistories)
	return scrapingHistories
}
