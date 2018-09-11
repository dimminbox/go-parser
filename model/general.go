package model

import (
	"encoding/json"
	"fmt"
	"os"
	//загрузка mysql драйвера
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

var DB *gorm.DB
var Connect *gorm.DB

const ATP_PLAYERS_URL = "https://www.atpworldtour.com"
const MYSCORE_PLAYERS_URL = "https://www.myscore.ru"
const ATP_PLAYERS_LIST_URL = "https://www.atpworldtour.com/en/rankings/singles?countryCode=all&rankRange="
const MYSCORE_PLAYERS_LIST_URL = "https://www.myscore.ru/tennis/rankings/atp/"

//Configuration - структура конфигурации микросервиса
type Configuration struct {
	Port      string
	Host      string
	User      string
	Password  string
	DbURI     string
	IsDebug   bool
	Resources string
}

var configuration Configuration

func init() {

	file, err := os.Open("config.json")
	if err != nil {
		panic(err)
	}

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&configuration)
	if err != nil {
		panic(err)
	}

}

//Init - инициализация конфигурации
func Init() Configuration {

	return configuration
}

//InitDB2 соединяется с БД и создаёт одно соединение
func InitDB() {

	if Connect == nil {

		db, err := gorm.Open("mysql", configuration.DbURI)
		db.LogMode(configuration.IsDebug)
		if err != nil {
			panic(fmt.Sprintf("Got error when connect database, the error is '%v'", err))
		}

		Connect = db
	}

}
