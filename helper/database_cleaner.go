package helper

import (
	"github.com/mass584/autotrader/entity"
	"gorm.io/gorm"
)

func DatabaseCleaner(db *gorm.DB) {
	db.Where("1 = 1").Delete(&entity.ScrapingHistory{})
	db.Where("1 = 1").Delete(&entity.Trade{})
}
