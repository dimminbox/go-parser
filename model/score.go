package model

type Score struct {
	ID       int     `gorm:"column:id;primary_key" json:"id"`
	Result   string  `gorm:"column:result"`
	Handicap float32 `gorm:"column:handicap"`
	Score    float32 `gorm:"column:score"`
}

// TableName sets the insert table name for this struct type
func (g *Score) TableName() string {
	return "score"
}
