package model

type Rating struct {
	ID         int    `gorm:"column:id;primary_key" json:"id"`
	Player     int    `form:"player" gorm:"column:player" json:"player" valid:"required"`
	Code       string `json:"-" gorm:"-"`
	Rating     int    `form:"rating" gorm:"column:rating" json:"rating" valid:"required"`
	DateUpdate string `gorm:"column:dateUpdate" json:"dateUpdate"`
	Points     int    `gorm:"column:points" json:"points"`
}

// TableName sets the insert table name for this struct type
func (g *Rating) TableName() string {
	return "rating"
}
