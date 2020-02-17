package model

import "time"

type GameWomenToday struct {
	ID          int       `gorm:"column:id;primary_key" json:"id"`
	Player1     int       `form:"player1" gorm:"column:player1" json:"player1"`
	PlayerCode1 string    `gorm:"-"`
	PlayerCode2 string    `gorm:"-"`
	Player2     int       `valid:"required" gorm:"column:player2" json:"player2"`
	DateEvent   time.Time `gorm:"column:dateEvent" json:"dateEvent"`
	OddAvg1     float64   `gorm:"column:oddAvg1" json:"OddAvg1"`
	OddAvg2     float64   `gorm:"column:oddAvg2" json:"OddAvg2"`
	OddAvgMy1   float32   `gorm:"column:oddAvgMy1" json:"oddAvgMy1"`
	OddAvgMy2   float32   `gorm:"column:oddAvgMy2" json:"oddAvgMy2"`
	URL         string    `gorm:"column:url" json:"url"`
	Tournir     string    `gorm:"column:tournir" json:"tournir"`
}

// TableName sets the insert table name for this struct type
func (g *GameWomenToday) TableName() string {
	return "todayWomen"
}
