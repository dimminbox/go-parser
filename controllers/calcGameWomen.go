package controller

import (
	"fmt"
	"os"
	"parser/model"
	"strconv"
	"strings"
)

var schema map[string]map[float32]float32

type OddLimit struct {
	Handi  float32
	Result float32
}

var schemaLimit map[string]map[string]OddLimit

func calcGame(today model.GameWomenToday) (oddAvgMy1 float32, oddAvgMy2 float32) {
	fmt.Printf("%+v\n", schema)
	os.Exit(1)
	oddAvgMy1 = 0
	oddAvgMy2 = 0

	//считаем рейтинг первого игрока
	var gamePlayer1 []model.WomenGame
	model.Connect.
		Where("player1 = ?", today.Player1).
		Or("player2 = ?", today.Player1).
		Order("dateEvent desc").
		Limit(20).
		Find(&gamePlayer1)

	for i, item := range gamePlayer1 {
		var rating model.WomenRating
		model.Connect.
			Where("player = ?", item.Player2).
			Where("dateUpdate <= ?", item.DateEvent).
			Order("dateUpdate desc").
			Limit(1).
			Find(&rating)
		if rating.Rating != 0 {
			gamePlayer1[i].Player2Rating = rating.Rating
		}

		model.Connect.
			Where("player = ?", item.Player1).
			Where("dateUpdate <= ?", item.DateEvent).
			Order("dateUpdate desc").
			Limit(1).
			Find(&rating)
		if rating.Rating != 0 {
			gamePlayer1[i].Player1Rating = rating.Rating
		}
	}

	count := 0
	for _, item := range gamePlayer1 {

		var prefix string
		if item.Player1 == today.Player1 {
			prefix = "win"
		} else {
			prefix = "lose"
		}
		if item.Player2Rating == 0 {
			continue
		}
		sets := strings.Split(item.Scores, ";")
		// если сыграно 2 сета
		if len(sets) == 2 {
			if len(sets[0]) == 2 {
				p1, _ := strconv.Atoi(string(sets[0][0]))
				p2, _ := strconv.Atoi(string(sets[0][1]))

				var dif float32
				if prefix == "win" {
					dif = 0 - (float32)(p1-p2) + 0.5
				} else {
					dif = (float32)(p1-p2) - 0.5
				}

				pref := fmt.Sprintf("%s2Set", prefix)
				if val, ok := schema[pref][dif]; ok {
					res := val / (float32)(item.Player2Rating)
					oddAvgMy1 += res
					count++
				} else {

					if prefix == "win" {

					}

					res := schema[prefix][0] / (float32)(item.Player2Rating)
					oddAvgMy1 += res
					count++
				}
			}
		}
		// если сыграно 3 сета
		if len(sets) == 3 {
			if len(sets[0]) == 2 {
				p1, _ := strconv.Atoi(string(sets[0][0]))
				p2, _ := strconv.Atoi(string(sets[0][1]))

				var dif float32
				if prefix == "win" {
					dif = 0 - (float32)(p1-p2) + 0.5
				} else {
					dif = (float32)(p1-p2) - 0.5
				}

				if val, ok := schema[prefix][dif]; ok {
					res := val / (float32)(item.Player2Rating)
					oddAvgMy1 += res
					count++
				} else {
					res := schema[prefix][0] / (float32)(item.Player2Rating)
					oddAvgMy1 += res
					count++
				}
			}
		}
		// если сыгран 1 сет
		if len(sets) == 1 {
			if len(sets[0]) == 2 {
				p1, _ := strconv.Atoi(string(sets[0][0]))
				p2, _ := strconv.Atoi(string(sets[0][1]))
				if p1 >= 6 {

					var dif float32
					if prefix == "win" {
						dif = 0 - (float32)(p1-p2) + 0.5
					} else {
						dif = (float32)(p1-p2) - 0.5
					}

					pref := fmt.Sprintf("%s2Set", prefix)
					if val, ok := schema[pref][dif]; ok {
						res := val / (float32)(item.Player2Rating)
						oddAvgMy1 += res
						count++
					} else {
						if dif < 0 {
							res := schema[prefix][0] / (float32)(item.Player2Rating)
							oddAvgMy1 += res
							count++
						}
					}
				}
			}
		}

		if count == 10 {
			break
		}
	}
	fmt.Println(oddAvgMy1)
	os.Exit(1)

	return
}

func getScoreByHandi(schema map[string]map[float32]float32) {

}
