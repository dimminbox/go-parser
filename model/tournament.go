package model

type Tournament struct {
	ID          int    `gorm:"column:id;primary_key" json:"id"`
	Name        string `gorm:"column:name" json:"name"`
	DateStart   string `gorm:"column:dateStart" json:"dateStart"`
	DateStop    string `gorm:"column:dateStop" json:"dateStop"`
	Myscore     string `gorm:"column:myscore" json:"myscore"`
	Tennisworld string `gorm:"column:tennisworld" json:"tennisworld"`
	Surface     string `gorm:"column:surface" json:"surface"`
	PrizeMoney  string `gorm:"column:prizeMoney" json:"prizeMoney"`
	Year        int    `gorm:"column:year" json:"year"`
	Type        string `gorm:"column:type" json:"type"`
	IsParse     int    `gorm:"column:isParse" json:"isParse"`
	Sgl         string `gorm:"column:sgl" json:"sgl"`
	Dbl         string `gorm:"column:dbl" json:"dbl"`
	Hospitality int    `gorm:"column:hospitality" json:"hospitality"`
}

// TableName sets the insert table name for this struct type
func (t *Tournament) TableName() string {
	return "tournament"
}
