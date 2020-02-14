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
		DATE   string = time.Now().Format("2006-01-02")
	)

	flag.StringVar(&METHOD, "m", METHOD, "метод парсинга (rating, player, game, gameWomenYear, gameWomenToday, tournament, all, women, gameWomenDay)")
	flag.IntVar(&COUNT, "count", COUNT, "количество записей парсинга")
	flag.IntVar(&YEAR, "year", YEAR, "год")
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
	case "gameWomenDay":
		t := time.Now().AddDate(0, 0, -1)
		controller.GameWomenDay(t.Year(), int(t.Month()), t.Day())
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
