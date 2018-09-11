package model

import (
	"database/sql"
)

type PointRule struct {
	ID          int            `gorm:"column:id;primary_key" json:"id"`
	Type        string         `gorm:"column:type" json:"type"`
	PrizeMoney  sql.NullString `gorm:"column:prizeMoney" json:"prizeMoney"`
	Players     sql.NullInt64  `gorm:"column:players" json:"players"`
	Round       sql.NullString `gorm:"column:round" json:"round"`
	Points      int            `gorm:"column:points" json:"points"`
	Hospitality int            `gorm:"column:hospitality" json:"hospitality"`
}

// TableName sets the insert table name for this struct type
func (p *PointRule) TableName() string {
	return "pointRule"
}
