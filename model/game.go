package model

type Game struct {
	ID                     int    `gorm:"column:id;primary_key" json:"id"`
	Player1                int    `form:"player1" gorm:"column:player1" json:"player1" valid:"required"`
	PlayerURL1             string `gorm:"-"`
	Status1                string `gorm:"column:status1" json:"status1"`
	PlayerURL2             string `gorm:"-"`
	Player2                int    `valid:"required" gorm:"column:player2" json:"player2"`
	Status2                string `gorm:"column:status2" json:"status2"`
	Winner                 int    `valid:"required" gorm:"column:winner" json:"winner"`
	DateEvent              string `gorm:"column:dateEvent" json:"dateEvent"`
	Scores                 string `gorm:"column:scores" json:"scores"`
	Stage                  string `valid:"required" gorm:"column:stage" json:"stage"`
	Aces1                  int    `gorm:"column:aces1" json:"aces1"`
	Aces2                  int    `gorm:"column:aces2" json:"aces2"`
	DoubleFaults1          int    `gorm:"column:doubleFaults1" json:"doubleFaults1"`
	DoubleFaults2          int    `gorm:"column:doubleFaults2" json:"doubleFaults2"`
	Serve1                 int    `gorm:"column:serve1" json:"serve1"`
	Serve2                 int    `gorm:"column:serve2" json:"serve2"`
	Serve1PointsWon1       int    `gorm:"column:serve1PointsWon1" json:"serve1PointsWon1"`
	Serve2PointsWon1       int    `gorm:"column:serve2PointsWon1" json:"serve2PointsWon1"`
	Serve2PointsWon2       int    `gorm:"column:serve2PointsWon2" json:"serve2PointsWon2"`
	Serve1PointsWon2       int    `gorm:"column:serve1PointsWon2" json:"serve1PointsWon2"`
	BreakPointsSaved1      int    `gorm:"column:breakPointsSaved1" json:"breakPointsSaved1"`
	BreakPointsSaved2      int    `gorm:"column:breakPointsSaved2" json:"breakPointsSaved2"`
	ServiceGamesPlayed1    int    `gorm:"column:serviceGamesPlayed1" json:"serviceGamesPlayed1"`
	ServiceGamesPlayed2    int    `gorm:"column:serviceGamesPlayed2" json:"serviceGamesPlayed2"`
	ReturnRating1          int    `gorm:"column:returnRating1" json:"returnRating1"`
	ReturnRating2          int    `gorm:"column:returnRating2" json:"returnRating2"`
	Serve1ReturnPointsWon1 int    `gorm:"column:serve1ReturnPointsWon1" json:"serve1ReturnPointsWon1"`
	Serve1ReturnPointsWon2 int    `gorm:"column:serve1ReturnPointsWon2" json:"serve1ReturnPointsWon2"`
	Serve2ReturnPointsWon1 int    `gorm:"column:serve2ReturnPointsWon1" json:"serve2ReturnPointsWon1"`
	Serve2ReturnPointsWon2 int    `gorm:"column:serve2ReturnPointsWon2" json:"serve2ReturnPointsWon2"`
	BreakPointsConverted1  int    `gorm:"column:breakPointsConverted1" json:"breakPointsConverted1"`
	BreakPointsConverted2  int    `gorm:"column:breakPointsConverted2" json:"breakPointsConverted2"`
	ReturnGamesPlayed1     int    `gorm:"column:returnGamesPlayed1" json:"returnGamesPlayed1"`
	ReturnGamesPlayed2     int    `gorm:"column:returnGamesPlayed2" json:"returnGamesPlayed2"`
	ServicePointsWon1      int    `gorm:"column:servicePointsWon1" json:"servicePointsWon1"`
	ServicePointsWon2      int    `gorm:"column:servicePointsWon2" json:"servicePointsWon2"`
	ReturnPointsWon1       int    `gorm:"column:returnPointsWon1" json:"returnPointsWon1"`
	ReturnPointsWon2       int    `gorm:"column:returnPointsWon2" json:"returnPointsWon2"`
	TotalPointsWon1        int    `gorm:"column:totalPointsWon1" json:"totalPointsWon1"`
	TotalPointsWon2        int    `gorm:"column:totalPointsWon2" json:"totalPointsWon2"`
	Duration               int    `gorm:"column:duration" json:"duration"`
	URL                    string `gorm:"column:url" json:"url"`
	Tournir                int    `valid:"required" gorm:"column:tournir" json:"tournir"`
}

// TableName sets the insert table name for this struct type
func (g *Game) TableName() string {
	return "game"
}
