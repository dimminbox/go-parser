package controller

import (
	"fmt"
	"log"
	"net/http"
	"parser/model"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/asaskevich/govalidator"
)

const GAME_URL = "https://www.atpworldtour.com"

func Games(year int) {

	var _players []model.Player
	players := map[string]int{}
	model.Connect.Find(&_players)

	for _, player := range _players {
		players[player.Code] = player.ID
	}

	var tournaments []model.Tournament

	games := []model.Game{}
	model.Connect.Where("year = ?", year).Find(&tournaments)

	//tournaments = tournaments[150:151]
	exGames := map[string]model.Game{}
	_exGames := []model.Game{}
	for _, tournament := range tournaments {

		model.Connect.Where("Tournir = ?", tournament.ID).
			Find(&_exGames)
		for _, exGame := range _exGames {
			exGames[exGame.URL] = exGame
		}

		ch := make(chan model.Game)
		games = parserGames(tournament)

		for i, game := range games {
			fmt.Println(game.URL)
			go parserGame(game, ch)

			if i%50 == 0 {
				fmt.Printf("%s\n", "pause")
				time.Sleep(1000 * time.Millisecond)
			} else {
				time.Sleep(800 * time.Millisecond)
			}
		}

		for i := 0; i < len(games); i++ {
			game := <-ch

			if _exGame, ok := exGames[game.URL]; ok {
				game.ID = _exGame.ID
			}

			hrefArr1 := strings.Split(game.PlayerURL1, "/")
			hrefArr2 := strings.Split(game.PlayerURL2, "/")

			var code1 string
			if len(hrefArr1) > 3 {
				code1 = hrefArr1[len(hrefArr1)-2]
			} else {
				code1 = ""
			}

			var code2 string
			if len(hrefArr2) > 3 {
				code2 = hrefArr2[len(hrefArr2)-2]
			} else {
				code2 = ""
			}

			if ID, ok := players[code1]; ok {
				game.Player1 = ID
			}

			if ID, ok := players[code2]; ok {
				game.Player2 = ID
			}

			_, err := govalidator.ValidateStruct(game)

			if err == nil {
				//fmt.Printf("%d - %s \n", game.ID, game.URL)
				model.Connect.Save(&game)
			} else {
				fmt.Printf("%s\n", game.PlayerURL1)
				fmt.Println(err)
			}

		}

	}

	time.Sleep(1000 * time.Millisecond)

}

func parserGames(tournament model.Tournament) (games []model.Game) {

	games = []model.Game{}

	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	client := &http.Client{Transport: tr}

	req, _ := http.NewRequest("GET", tournament.Tennisworld, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/44.0.2403.157 Safari/537.36")
	res, err := client.Do(req)
	if err == nil {
		defer res.Body.Close()
		if res.StatusCode == 200 {

			doc, err2 := goquery.NewDocumentFromReader(res.Body)
			if err2 == nil {

				doc.Find("table.day-table tbody tr").Each(func(i int, s *goquery.Selection) {
					game := model.Game{Tournir: tournament.ID, Winner: 1}
					s.Find("td.day-table-seed span").Each(func(i int, s *goquery.Selection) {

						r := strings.NewReplacer("(", "", ")", "", " ", "")
						if i == 0 {
							game.Status1 = r.Replace(strings.TrimSpace(s.Text()))
						} else if i == 1 {
							game.Status2 = r.Replace(strings.TrimSpace(s.Text()))
						}
					})

					s.Find("td.day-table-score a").Each(func(i int, s *goquery.Selection) {
						href, _ := s.Attr("href")
						game.URL = GAME_URL + href

						score := strings.Replace(strings.TrimSpace(s.Text()), " ", ";", -1)
						game.Scores = score

					})

					games = append(games, game)

				})
			} else {
				log.Println(err2)
			}

		}
	} else {
		log.Println(err)
	}
	return
}
func parserGame(game model.Game, ch chan model.Game) {

	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	client := &http.Client{Transport: tr}
	req, _ := http.NewRequest("GET", game.URL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/44.0.2403.157 Safari/537.36")
	res, err := client.Do(req)
	if err == nil {
		defer res.Body.Close()
		if res.StatusCode == 200 {
			doc, err1 := goquery.NewDocumentFromReader(res.Body)
			if err1 == nil {

				doc.Find("table.scores-table caption.match-title span.title-area").Each(func(i int, s *goquery.Selection) {
					game.Stage = strings.TrimSpace(s.Text())
				})

				doc.Find("tr.match-info-row > td.time").Each(func(i int, s *goquery.Selection) {
					durArr := strings.Split(s.Text(), ":")
					if len(durArr) == 3 {
						hours, _ := strconv.Atoi(strings.TrimSpace(durArr[0]))
						minutes, _ := strconv.Atoi(durArr[1])
						duration := hours*60 + minutes
						game.Duration = duration
					}
				})

				r := strings.NewReplacer("es", "en", "de", "en", "pt", "en")
				doc.Find("div.player-left-name a").Each(func(i int, s *goquery.Selection) {
					game.PlayerURL1, _ = s.Attr("href")
					game.PlayerURL1 = GAME_URL + r.Replace(game.PlayerURL1)
				})

				doc.Find("div.player-right-name a").Each(func(i int, s *goquery.Selection) {
					game.PlayerURL2, _ = s.Attr("href")
					game.PlayerURL2 = GAME_URL + r.Replace(game.PlayerURL2)
				})

				doc.Find("tr.match-stats-row").Eq(1).Find("td.match-stats-number-left").Each(func(i int, s *goquery.Selection) {
					game.Aces1, _ = strconv.Atoi(strings.TrimSpace(s.Text()))
				})

				doc.Find("tr.match-stats-row").Eq(1).Find("td.match-stats-number-right").Each(func(i int, s *goquery.Selection) {
					game.Aces2, _ = strconv.Atoi(strings.TrimSpace(s.Text()))
				})

				doc.Find("tr.match-stats-row").Eq(2).Find("td.match-stats-number-left").Each(func(i int, s *goquery.Selection) {
					game.DoubleFaults1, _ = strconv.Atoi(strings.TrimSpace(s.Text()))
				})

				doc.Find("tr.match-stats-row").Eq(2).Find("td.match-stats-number-right").Each(func(i int, s *goquery.Selection) {
					game.DoubleFaults2, _ = strconv.Atoi(strings.TrimSpace(s.Text()))
				})

				doc.Find("tr.match-stats-row").Eq(3).Find("td.match-stats-number-left span").Eq(0).Each(func(i int, s *goquery.Selection) {
					game.Serve1, _ = strconv.Atoi(strings.TrimSpace(strings.Replace(s.Text(), "%", "", -1)))
				})

				doc.Find("tr.match-stats-row").Eq(3).Find("td.match-stats-number-right span").Eq(0).Each(func(i int, s *goquery.Selection) {
					game.Serve2, _ = strconv.Atoi(strings.TrimSpace(strings.Replace(s.Text(), "%", "", -1)))
				})

				doc.Find("tr.match-stats-row").Eq(4).Find("td.match-stats-number-left span").Eq(0).Each(func(i int, s *goquery.Selection) {
					game.Serve1PointsWon1, _ = strconv.Atoi(strings.TrimSpace(strings.Replace(s.Text(), "%", "", -1)))
				})

				doc.Find("tr.match-stats-row").Eq(4).Find("td.match-stats-number-right span").Eq(0).Each(func(i int, s *goquery.Selection) {
					game.Serve1PointsWon2, _ = strconv.Atoi(strings.TrimSpace(strings.Replace(s.Text(), "%", "", -1)))
				})

				doc.Find("tr.match-stats-row").Eq(5).Find("td.match-stats-number-left span").Eq(0).Each(func(i int, s *goquery.Selection) {
					game.Serve2PointsWon1, _ = strconv.Atoi(strings.TrimSpace(strings.Replace(s.Text(), "%", "", -1)))
				})

				doc.Find("tr.match-stats-row").Eq(5).Find("td.match-stats-number-right span").Eq(0).Each(func(i int, s *goquery.Selection) {
					game.Serve2PointsWon2, _ = strconv.Atoi(strings.TrimSpace(strings.Replace(s.Text(), "%", "", -1)))
				})

				doc.Find("tr.match-stats-row").Eq(6).Find("td.match-stats-number-left span").Eq(0).Each(func(i int, s *goquery.Selection) {
					game.BreakPointsSaved1, _ = strconv.Atoi(strings.TrimSpace(strings.Replace(s.Text(), "%", "", -1)))
				})

				doc.Find("tr.match-stats-row").Eq(6).Find("td.match-stats-number-right span").Eq(0).Each(func(i int, s *goquery.Selection) {
					game.BreakPointsSaved2, _ = strconv.Atoi(strings.TrimSpace(strings.Replace(s.Text(), "%", "", -1)))
				})

				doc.Find("tr.match-stats-row").Eq(6).Find("td.match-stats-number-left span").Eq(0).Each(func(i int, s *goquery.Selection) {
					game.BreakPointsSaved1, _ = strconv.Atoi(strings.TrimSpace(strings.Replace(s.Text(), "%", "", -1)))
				})

				doc.Find("tr.match-stats-row").Eq(7).Find("td.match-stats-number-left span").Eq(0).Each(func(i int, s *goquery.Selection) {
					game.ServiceGamesPlayed1, _ = strconv.Atoi(strings.TrimSpace(strings.Replace(s.Text(), "%", "", -1)))
				})

				doc.Find("tr.match-stats-row").Eq(7).Find("td.match-stats-number-right span").Eq(0).Each(func(i int, s *goquery.Selection) {
					game.ServiceGamesPlayed2, _ = strconv.Atoi(strings.TrimSpace(strings.Replace(s.Text(), "%", "", -1)))
				})

				doc.Find("tr.match-stats-row").Eq(8).Find("td.match-stats-number-left span").Eq(0).Each(func(i int, s *goquery.Selection) {
					game.ReturnRating1, _ = strconv.Atoi(strings.TrimSpace(strings.Replace(s.Text(), "%", "", -1)))
				})

				doc.Find("tr.match-stats-row").Eq(8).Find("td.match-stats-number-right span").Eq(0).Each(func(i int, s *goquery.Selection) {
					game.ReturnRating2, _ = strconv.Atoi(strings.TrimSpace(strings.Replace(s.Text(), "%", "", -1)))
				})

				doc.Find("tr.match-stats-row").Eq(9).Find("td.match-stats-number-left span").Eq(0).Each(func(i int, s *goquery.Selection) {
					game.Serve1ReturnPointsWon1, _ = strconv.Atoi(strings.TrimSpace(strings.Replace(s.Text(), "%", "", -1)))
				})

				doc.Find("tr.match-stats-row").Eq(9).Find("td.match-stats-number-right span").Eq(0).Each(func(i int, s *goquery.Selection) {
					game.Serve1ReturnPointsWon2, _ = strconv.Atoi(strings.TrimSpace(strings.Replace(s.Text(), "%", "", -1)))
				})

				doc.Find("tr.match-stats-row").Eq(10).Find("td.match-stats-number-left span").Eq(0).Each(func(i int, s *goquery.Selection) {
					game.Serve2ReturnPointsWon1, _ = strconv.Atoi(strings.TrimSpace(strings.Replace(s.Text(), "%", "", -1)))
				})

				doc.Find("tr.match-stats-row").Eq(10).Find("td.match-stats-number-right span").Eq(0).Each(func(i int, s *goquery.Selection) {
					game.Serve2ReturnPointsWon2, _ = strconv.Atoi(strings.TrimSpace(strings.Replace(s.Text(), "%", "", -1)))
				})

				doc.Find("tr.match-stats-row").Eq(11).Find("td.match-stats-number-left span").Eq(0).Each(func(i int, s *goquery.Selection) {
					game.BreakPointsConverted1, _ = strconv.Atoi(strings.TrimSpace(strings.Replace(s.Text(), "%", "", -1)))
				})

				doc.Find("tr.match-stats-row").Eq(11).Find("td.match-stats-number-right span").Eq(0).Each(func(i int, s *goquery.Selection) {
					game.BreakPointsConverted2, _ = strconv.Atoi(strings.TrimSpace(strings.Replace(s.Text(), "%", "", -1)))
				})

				doc.Find("tr.match-stats-row").Eq(12).Find("td.match-stats-number-left span").Eq(0).Each(func(i int, s *goquery.Selection) {
					game.ReturnGamesPlayed1, _ = strconv.Atoi(strings.TrimSpace(strings.Replace(s.Text(), "%", "", -1)))
				})

				doc.Find("tr.match-stats-row").Eq(12).Find("td.match-stats-number-right span").Eq(0).Each(func(i int, s *goquery.Selection) {
					game.ReturnGamesPlayed2, _ = strconv.Atoi(strings.TrimSpace(strings.Replace(s.Text(), "%", "", -1)))
				})

				doc.Find("tr.match-stats-row").Eq(13).Find("td.match-stats-number-left span").Eq(0).Each(func(i int, s *goquery.Selection) {
					game.ServicePointsWon1, _ = strconv.Atoi(strings.TrimSpace(strings.Replace(s.Text(), "%", "", -1)))
				})

				doc.Find("tr.match-stats-row").Eq(13).Find("td.match-stats-number-right span").Eq(0).Each(func(i int, s *goquery.Selection) {
					game.ServicePointsWon2, _ = strconv.Atoi(strings.TrimSpace(strings.Replace(s.Text(), "%", "", -1)))
				})

				doc.Find("tr.match-stats-row").Eq(14).Find("td.match-stats-number-left span").Eq(0).Each(func(i int, s *goquery.Selection) {
					game.ReturnPointsWon1, _ = strconv.Atoi(strings.TrimSpace(strings.Replace(s.Text(), "%", "", -1)))
				})

				doc.Find("tr.match-stats-row").Eq(14).Find("td.match-stats-number-right span").Eq(0).Each(func(i int, s *goquery.Selection) {
					game.ReturnPointsWon2, _ = strconv.Atoi(strings.TrimSpace(strings.Replace(s.Text(), "%", "", -1)))
				})

				doc.Find("tr.match-stats-row").Eq(15).Find("td.match-stats-number-left span").Eq(0).Each(func(i int, s *goquery.Selection) {
					game.TotalPointsWon1, _ = strconv.Atoi(strings.TrimSpace(strings.Replace(s.Text(), "%", "", -1)))
				})

				doc.Find("tr.match-stats-row").Eq(15).Find("td.match-stats-number-right span").Eq(0).Each(func(i int, s *goquery.Selection) {
					game.TotalPointsWon2, _ = strconv.Atoi(strings.TrimSpace(strings.Replace(s.Text(), "%", "", -1)))
				})

			} else {
				log.Println(err1)
			}

		}
	} else {
		log.Println(err)
	}
	ch <- game

}
