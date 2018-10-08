package controller

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"parser/model"
	"regexp"
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
			//fmt.Println(game.URL)
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

			if ID1, ok1 := players[code1]; ok1 {
				game.Player1 = ID1
			}

			if ID2, ok2 := players[code2]; ok2 {
				game.Player2 = ID2
			}

			_, err := govalidator.ValidateStruct(game)

			if err == nil {
				//fmt.Printf("%d - %s \n", game.ID, game.URL)
				model.Connect.Save(&game)
			} else {
				fmt.Printf("Code 1 %s\n", code1)
				fmt.Printf("Code 2 %s\n", code2)
				fmt.Printf("Game URL %s\n", game.URL)
				fmt.Println(err)
			}

		}

	}

	time.Sleep(1000 * time.Millisecond)

}

func parserGamesMsc(res *http.Response) {

	type Match struct {
		Url     string
		Player1 string
		Player2 string
		Date    time.Time
	}

	matches := map[string][]Match{}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err == nil {

		doc.Find("div#tournament-page-data-results").Each(func(i int, s *goquery.Selection) {
			body := s.Text()
			chunks := strings.Split(body, "~AA÷")

			r1, _ := regexp.Compile(`¬ER÷(.+)¬RW÷`)
			r2, _ := regexp.Compile(`(.+)¬AD÷`)

			for _, chunk := range chunks {

				stages := r1.FindStringSubmatch(chunk)
				urls := r2.FindStringSubmatch(chunk)

				if (len(stages) == 2) && (len(urls) == 2) {
					url := "https://www.myscore.ru/match/" + urls[1] + "/"
					stage := stages[1]

					if _, ok := matches[stage]; ok {
						matches[stage] = append(matches[stage], Match{Url: url})
					} else {
						matches[stage] = make([]Match, 0, 0)
						matches[stage] = append(matches[stage], Match{Url: url})
					}
				}

			}

		})

		client := &http.Client{}

		r3, _ := regexp.Compile(`window\.open\(\'(.+)\'\)`)
		r4, _ := regexp.Compile(`g2utime\s\=\s(\d+)\;`)

		for stage, _matches := range matches {

			for index, match := range _matches {

				req, _ := http.NewRequest("GET", match.Url, nil)
				req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/44.0.2403.157 Safari/537.36")
				res, err := client.Do(req)
				if err == nil {
					defer res.Body.Close()
					if res.StatusCode == 200 {

						doc, err2 := goquery.NewDocumentFromReader(res.Body)

						if err2 == nil {
							doc.Find("div.tname__text a").Each(func(i int, s *goquery.Selection) {
								href, _ := s.Attr("onclick")

								playerUrl := r3.FindStringSubmatch(href)

								if len(playerUrl) == 2 {

									if i == 0 {
										matches[stage][index].Player1 = "https://www.myscore.ru" + playerUrl[1]
									}

									if i == 1 {
										matches[stage][index].Player2 = "https://www.myscore.ru" + playerUrl[1]
									}

								}

							})

						}
						html, _ := doc.Html()
						matchTime := r4.FindStringSubmatch(html)
						if len(matchTime) == 2 {

							_date, err := strconv.ParseInt(matchTime[1], 10, 64)
							if err != nil {
								panic(err)
							}

							matches[stage][index].Date = time.Unix(_date, 0)
						}
					}
				}
			}

		}

		/*doc.Find("table.tennis tbody tr.stage-finished").Each(func(i int, s *goquery.Selection) {
			id, _ := s.Attr("id")
			id = strings.Replace(id, "g_2_", "", -1)
			matches[stage] = append(matches[stage], "https://www.myscore.ru/match/"+id+"/")
		})*/

	}
	fmt.Println(matches)
	os.Exit(1)
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

				reqMsc, _ := http.NewRequest("GET", tournament.Myscore, nil)
				resMsc, errMsc := client.Do(reqMsc)
				if errMsc == nil {
					defer resMsc.Body.Close()
					if resMsc.StatusCode == 200 {

						parserGamesMsc(resMsc)

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

					}
				}
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
