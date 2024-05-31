package database

import (
	"github.com/mass584/autotrader/entity"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func SaveScrapingHistory(
	db *gorm.DB,
	scrapingHistory entity.ScrapingHistory,
) (*entity.ScrapingHistory, error) {
	result := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"scraping_status"}),
	}).Create(&scrapingHistory)

	if result.Error != nil {
		return nil, errors.WithStack(result.Error)
	}

	return &scrapingHistory, nil
}

func GetScrapingHistoriesByStatus(
	db *gorm.DB,
	exchange_place entity.ExchangePlace,
	exchange_pair entity.ExchangePair,
	status entity.ScrapingStatus,
) ([]entity.ScrapingHistory, error) {
	var scrapingHistories []entity.ScrapingHistory
	result := db.
		Where("exchange_place = ?", exchange_place).
		Where("exchange_pair = ?", exchange_pair).
		Where("scraping_status = ?", status).
		Order("from_id DESC").
		Find(&scrapingHistories)

	if result.Error != nil {
		return nil, errors.WithStack(result.Error)
	}
	return scrapingHistories, nil
}
