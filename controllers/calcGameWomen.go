package controller

import (
	"fmt"
	"math"
	"parser/model"
	"strconv"
	"strings"
	"time"
)

var schema map[string]map[float32]float32

type OddLimit struct {
	Handi  float32
	Result float32
}

var schemaLimit map[string]map[string]OddLimit

func CalcGame(id int) {

	var today model.GameWomenToday
	model.Connect.Where("id = ?", id).Find(&today)
	if today.ID == 0 {
		var game model.WomenGame
		model.Connect.Where("id = ?", id).Find(&game)
		if game.ID != 0 {
			today = model.GameWomenToday{
				ID:        id,
				Player1:   game.Player1,
				Player2:   game.Player2,
				DateEvent: game.DateEvent,
			}
		}
	}

	if today.ID != 0 {
		calcGame(today)
	} else {
		fmt.Println("Игры не найдено.")
	}

}
func calcGame(today model.GameWomenToday) (oddAvgMy1 float32, oddAvgMy2 float32, cnt1 int, cnt2 int) {

	getSchema()
	fmt.Printf("Рассчёт игры %d\n", today.ID)
	var player1 model.Women
	model.Connect.Where("id = ?", today.Player1).Find(&player1)
	fmt.Println(player1.Name)
	res1, cnt1 := CalcPlayer(today.Player1)

	fmt.Println("")
	var player2 model.Women
	model.Connect.Where("id = ?", today.Player2).Find(&player2)
	fmt.Println(player2.Name)
	res2, cnt2 := CalcPlayer(today.Player2)

	oddAvgMy1 = (float32)(math.Round((float64)((res1+res2)/res1*100)) / 100)
	oddAvgMy2 = (float32)(math.Round((float64)((res1+res2)/res2*100)) / 100)
	fmt.Printf("\n%s - %0.2f cnt1 %d\n", player1.Name, oddAvgMy1, cnt1)
	fmt.Printf("\n%s - %0.2f cnt2 %d\n", player2.Name, oddAvgMy2, cnt2)
	return
}

func CalcPlayer(player int) (oddAvgMy float32, count int) {

	oddAvgMy = 0

	//считаем рейтинг первого игрока
	var gamePlayer1 []model.WomenGame
	model.Connect.
		Where("player1 = ?", player).
		Or("player2 = ?", player).
		Order("dateEvent desc").
		Limit(20).
		Find(&gamePlayer1)

	for i, item := range gamePlayer1 {

		var p1 model.Women
		model.Connect.Where("id = ?", item.Player1).Find(&p1)
		gamePlayer1[i].Player1Name = p1.Name

		var p2 model.Women
		model.Connect.Where("id = ?", item.Player2).Find(&p2)
		gamePlayer1[i].Player2Name = p2.Name

		var rating2 model.WomenRating
		model.Connect.
			Where("player = ?", item.Player2).
			Where("dateUpdate <= ?", item.DateEvent).
			Order("dateUpdate desc").
			Order("id desc").
			Limit(1).
			Find(&rating2)
		if rating2.Rating != 0 {
			gamePlayer1[i].Player2Rating = rating2.Rating
			gamePlayer1[i].Player2RatingDate = rating2.DateUpdate
		}

		var rating1 model.WomenRating
		model.Connect.
			Where("player = ?", item.Player1).
			Where("dateUpdate <= ?", item.DateEvent).
			Order("dateUpdate desc").
			Order("id desc").
			Limit(1).
			Find(&rating1)
		if rating1.Rating != 0 {
			gamePlayer1[i].Player1Rating = rating1.Rating
			gamePlayer1[i].Player1RatingDate = rating1.DateUpdate
		}
	}

	for _, item := range gamePlayer1 {

		var prefix string
		if item.Player1 == player {
			prefix = "win"
		} else {
			prefix = "lose"
		}
		if item.Player1Rating == 0 || item.Player2Rating == 0 {
			continue
		}
		_sets := strings.Split(item.Scores, ";")
		var sets []string
		for _, _set := range _sets {
			if _set != "" {
				sets = append(sets, strings.Replace(_set, ":", "", -1))
			}
		}
		if len(sets) == 2 && prefix == "win" {
			prefix = "win2Set"
		}
		result, flag := getValByGame(sets, prefix)
		if flag {
			var rating int
			var ratingDate time.Time
			if strings.HasPrefix(prefix, "win") {
				rating = item.Player2Rating
				ratingDate = item.Player2RatingDate
			} else {
				rating = item.Player1Rating
				ratingDate = item.Player1RatingDate
			}
			res := result / (float32)(rating)
			oddAvgMy += res
			count++
			fmt.Printf("%s - %s, score %s, result %0.2f, value %0.7f, rating %d, ratingDate %+v \n", item.Player1Name, item.Player2Name, item.Scores, result, res, rating, ratingDate)
		}

		if count == 10 {
			break
		}
	}

	return
}

func getValByGame(sets []string, prefix string) (result float32, flag bool) {

	var dif float32
	if len(sets) == 0 {
		flag = false
		return
	}
	// если не доигран даже один сет то игру не берём в рассчтё
	if len(sets) == 1 {
		if len(sets[0]) == 2 {
			p1, _ := strconv.Atoi(string(sets[0][0]))
			p2, _ := strconv.Atoi(string(sets[0][1]))
			if p1 < 6 {
				flag = false
				return
			} else {
				dif = (float32)(p1 - p2)
			}

		}
	} else {
		for _, set := range sets {

			if len(set) < 2 {
				flag = false
				return
			}
			p1, _ := strconv.Atoi(string(set[0]))
			p2, _ := strconv.Atoi(string(set[1]))
			dif += (float32)(p1 - p2)
		}
	}

	if dif < 0 {
		dif *= -1
	}

	flag = true
	if prefix == "win" {
		dif = dif - 0.5
	} else if prefix == "win2Set" {
		dif = dif - 0.5
	} else {
		dif = dif + 0.5
	}

	if val, ok := schema[prefix][dif]; ok {
		result = val
	} else {
		max := schemaLimit[prefix]["max"]
		min := schemaLimit[prefix]["min"]
		if dif > max.Handi {
			if strings.HasPrefix(prefix, "lose") {
				result = (float32)(math.Round((float64)(max.Result)))
			} else {
				result = max.Result
			}
		}
		if dif < min.Handi {
			if strings.HasPrefix(prefix, "win") {
				result = min.Result
			} else {
				result = (float32)(math.Round((float64)(min.Result)))
			}
		}
	}
	fmt.Println("\ndifference ", dif)
	/*fmt.Printf("%+v\n", schemaLimit)
	fmt.Println(dif)
	fmt.Println(prefix)
	fmt.Println(result)
	os.Exit(1)*/
	return
}
