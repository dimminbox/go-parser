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
	)

	flag.StringVar(&METHOD, "m", METHOD, "метод парсинга (rating, player, game, tournament, all)")
	flag.IntVar(&COUNT, "count", COUNT, "количество записей парсинга")
	flag.IntVar(&YEAR, "year", YEAR, "год")
	flag.Parse()

	switch METHOD {
	case "player":
		controller.Players(COUNT)
	case "tournament":
		controller.Tournaments(YEAR)
	case "game":
		controller.Games(YEAR)
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
