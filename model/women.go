package model

type Women struct {
	ID             int    `gorm:"column:id;primary_key" json:"id"`
	Name           string `gorm:"column:name" json:"name"`
	Country        string `gorm:"column:country" json:"country"`
	Photo          string `gorm:"column:photo" json:"photo"`
	Age            string `gorm:"column:age" json:"age"`
	Pro            string `gorm:"column:pro" json:"pro"`
	Hand           string `gorm:"column:hand" json:"hand"`
	Tennisexplorer string `gorm:"column:tennisexplorer" json:"tennisexplorer"`
	Myscore    		string `gorm:"column:myscore" json:"myscore"`
	Sex            int    `gorm:"column:sex" json:"sex"`
	Coutry         string `gorm:"column:coutry" json:"coutry"`
	TurnirPro      int    `gorm:"column:turnirPro" json:"turnirPro"`
	Weight         string `gorm:"column:weight" json:"weight"`
	Height         string `gorm:"column:height" json:"height"`
	Birthplace     string `gorm:"column:birthplace" json:"birthplace"`
	Coach          string `gorm:"column:coach" json:"coach"`
	Rank           int    `gorm:"column:rank" json:"rank"`
	MoveRank       int    `gorm:"column:moveRank" json:"moveRank"`
	WinCY          int    `gorm:"column:winCY" json:"winCY"`
	LoseCY         int    `gorm:"column:loseCY" json:"loseCY"`
	PrizeMoneyCY   int    `gorm:"column:prizeMoneyCY" json:"prizeMoneyCY"`
	HighRank       int    `gorm:"column:highRank" json:"highRank"`
	Win            int    `gorm:"column:win" json:"win"`
	Lose           int    `gorm:"column:lose" json:"lose"`
	PrizeMoney     int    `gorm:"column:prizeMoney" json:"prizeMoney"`
	Points         int    `gorm:"column:points" json:"points"`
	Titles         int    `gorm:"column:titles" json:"titles"`
	TitlesCY       int    `gorm:"column:titlesCY" json:"titlesCY"`
	Code           string `gorm:"column:code" json:"code"`
}

// TableName sets the insert table name for this struct type
func (p *Women) TableName() string {
	return "women"
}

func (p *Women) Women(ID int) {

	if ID > 0 {
		Connect.Model(p).Update()
	}
}
