package database

import (
	"github.com/mass584/autotrader/entity"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func SavePosition(db *gorm.DB, position entity.Position) (*entity.Position, error) {
	result := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"position_status", "sell_price", "sell_time"}),
	}).Create(&position)

	if result.Error != nil {
		return nil, errors.WithStack(result.Error)
	}

	return &position, nil
}

func GetPositionsByStatus(
	db *gorm.DB,
	exchange_place entity.ExchangePlace,
	exchange_pair entity.ExchangePair,
	position_type entity.PositionType,
	position_status entity.PositionStatus,
) ([]entity.Position, error) {
	var positions []entity.Position
	result := db.
		Where("exchange_place = ?", exchange_place).
		Where("exchange_pair = ?", exchange_pair).
		Where("position_type = ?", position_type).
		Where("position_status = ?", position_status).
		Find(&positions)

	if result.Error != nil {
		return nil, errors.WithStack(result.Error)
	}

	return positions, nil
}
