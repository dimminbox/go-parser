package main

import (
	"flag"
	controller "parser/controllers"
	"parser/model"
)

func main() {

	model.InitDB()
	var (
		METHOD string = "all"
		COUNT  int    = 10
	)

	flag.StringVar(&METHOD, "m", METHOD, "метод парсинга (player, game, tournir, all)")
	flag.IntVar(&COUNT, "count", COUNT, "количество записей парсинга")
	flag.Parse()
	if (METHOD == "all") || (METHOD == "player") {
		controller.Players(COUNT)
	}

}
