package main

import (
	"flag"
	controller "parser/controllers"
	"parser/model"
	"time"
)

func main() {

	model.InitDB()
	var (
		METHOD string = "all"
		COUNT  int    = 10
		YEAR   int    = time.Now().Year()
		ID     int    = 0
		DATE   string = time.Now().Format("2006-01-02")
	)

	flag.StringVar(&METHOD, "m", METHOD, "метод парсинга (rating, player, game, gameWomenYear, gameWomenToday, tournament, all, women, gameWomenDay, calcGame)")
	flag.IntVar(&COUNT, "count", COUNT, "количество записей парсинга")
	flag.IntVar(&YEAR, "year", YEAR, "год")
	flag.IntVar(&ID, "id", ID, "id")
	flag.StringVar(&DATE, "date", DATE, "дата в формате 2006-01-02")
	flag.Parse()

	switch METHOD {
	case "player":
		controller.Players(COUNT)
	case "game":
		controller.Games(YEAR)
	case "women":
		controller.Womens(DATE)
	case "gameWomenYear":
		controller.GameWomenYear(YEAR)
	case "calcGame":
		controller.CalcGame(ID)
	case "gameWomenDay":
		controller.GameWomenDay(DATE)
	case "gameWomenToday":
		controller.GameWomenToday(DATE)
	case "tournament":
		controller.Tournaments(YEAR)
	case "rating":
		controller.GetRating()
	case "all":
		YEAR = 2017
		controller.GetRating()
		controller.Players(COUNT)
		controller.Tournaments(YEAR)
		controller.Games(YEAR)
	}

}
