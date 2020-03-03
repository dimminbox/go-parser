package controller

import (
	"fmt"
	"math"
	"net/http"
	"os"
	"parser/model"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func GameWomenDay(date string) {

	t, _ := time.Parse("2006-01-02", date)
	year := t.Year()
	month := int(t.Month())
	day := t.Day()

	var _exPlayers []model.Women
	model.Connect.Find(&_exPlayers)

	exPlayers := map[string]model.Women{}
	for _, player := range _exPlayers {
		exPlayers[player.Code] = player
	}

	games := parserGamesWomenDay(year, month, day)
	for _, item := range games {
		item.Player1 = exPlayers[item.PlayerCode1].ID
		item.Player2 = exPlayers[item.PlayerCode2].ID
		model.Connect.Save(&item)
		fmt.Printf("%+v\n", item)
	}
}

func getSchema() {

	schema = map[string]map[float32]float32{}
	schemaLimit = map[string]map[string]OddLimit{
		"win": map[string]OddLimit{
			"max": OddLimit{},
			"min": OddLimit{},
		},
		"win2Set": map[string]OddLimit{
			"max": OddLimit{},
			"min": OddLimit{},
		},
		"lose": map[string]OddLimit{
			"max": OddLimit{},
			"min": OddLimit{},
		},
	}
	schema["win"] = map[float32]float32{}
	schema["win2Set"] = map[float32]float32{}
	schema["lose"] = map[float32]float32{}

	var scores []model.Score
	model.Connect.Find(&scores)
	for _, item := range scores {
		if item.Result == "Win" {

			if item.Handicap > schemaLimit["win"]["max"].Handi {
				t := schemaLimit["win"]["max"]
				t.Handi = item.Handicap
				t.Result = item.Score
				schemaLimit["win"]["max"] = t
			}

			if item.Handicap < schemaLimit["win"]["max"].Handi {
				t := schemaLimit["win"]["min"]
				t.Handi = item.Handicap
				t.Result = item.Score
				schemaLimit["win"]["min"] = t
			}
			schema["win"][item.Handicap] = item.Score
		}
		if strings.HasPrefix(item.Result, "Win2set") {
			if item.Handicap > schemaLimit["win2Set"]["max"].Handi {
				t := schemaLimit["win2Set"]["max"]
				t.Handi = item.Handicap
				t.Result = item.Score
				schemaLimit["win2Set"]["max"] = t
			}

			if item.Handicap < schemaLimit["win2Set"]["max"].Handi {
				t := schemaLimit["win2Set"]["min"]
				t.Handi = item.Handicap
				t.Result = item.Score
				schemaLimit["win2Set"]["min"] = t
			}
			schema["win2Set"][item.Handicap] = item.Score
		}

		if strings.HasPrefix(item.Result, "Lose") {
			if item.Handicap > schemaLimit["lose"]["max"].Handi {
				t := schemaLimit["lose"]["max"]
				t.Handi = item.Handicap
				t.Result = item.Score
				schemaLimit["lose"]["max"] = t
			}

			if item.Handicap <= schemaLimit["lose"]["min"].Handi {
				t := schemaLimit["lose"]["min"]
				t.Handi = item.Handicap
				t.Result = item.Score
				schemaLimit["lose"]["min"] = t
			}
			schema["lose"][item.Handicap] = item.Score
		}
	}
}
func GameWomenToday(date string) {

	getSchema()
	model.Connect.Delete(model.GameWomenToday{})

	t, _ := time.Parse("2006-01-02", date)

	games := []model.GameWomenToday{}
	month := int(t.Month())
	year := t.Year()
	day := t.Day()

	var _month string
	if month > 9 {
		_month = fmt.Sprintf("%d", month)
	} else {
		_month = fmt.Sprintf("0%d", month)
	}

	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	client := &http.Client{Transport: tr}
	req, _ := http.NewRequest("GET", fmt.Sprintf("https://www.tennisexplorer.com/matches/?type=wta-single&year=%d&month=%s&day=%d", year, _month, day), nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/44.0.2403.157 Safari/537.36")
	res, err := client.Do(req)

	if err == nil {
		defer res.Body.Close()
		if res.StatusCode == 200 {
			doc, err2 := goquery.NewDocumentFromReader(res.Body)
			if err2 == nil {
				var tournir string
				var id int
				var matchURL string
				var game model.GameWomenToday
				var isFirst bool = true

				doc.Find("table.result > tbody > tr").Each(func(i int, s *goquery.Selection) {
					class, _ := s.Attr("class")
					if class == "head flags" {
						tournir = s.Find("td.t-name > a").Text()
					} else {
						tmpScore := []int{}
						s.Find("td.score").Each(func(j int, q *goquery.Selection) {
							val, _ := strconv.Atoi(string(strings.Trim(q.Text(), "&nbsp;")))
							tmpScore = append(tmpScore, val)
						})

						s.Find("td:last-child > a").Each(func(i int, q *goquery.Selection) {
							matchURL, _ = q.Attr("href")
							fmt.Sscanf(matchURL, "/match-detail/?id=%d", &id)
						})
						//fmt.Printf("%+v %s \n", tmpScore, matchURL)

						href, _ := s.Find("td.t-name > a").Attr("href")
						chunks := strings.Split(href, "/")

						if len(chunks) == 4 {
							if isFirst {

								game = model.GameWomenToday{
									DateEvent:   time.Now(),
									ID:          id,
									URL:         matchURL,
									Tournir:     tournir,
									PlayerCode1: chunks[2],
								}
								s.Find("td.course").Each(func(j int, q *goquery.Selection) {
									if j == 0 {
										odd, _ := strconv.ParseFloat(q.Text(), 32)
										game.OddAvg1 = math.Round(odd*100) / 100
									}
									if j == 1 {
										odd, _ := strconv.ParseFloat(q.Text(), 32)
										game.OddAvg2 = math.Round(odd*100) / 100
									}
								})
								isFirst = false
							} else {
								game.PlayerCode2 = chunks[2]
								if len(tmpScore) == 5 {
									if tmpScore[0] == 0 && tmpScore[1] == 0 && game.OddAvg1 != 0 && game.OddAvg2 != 0 {
										games = append(games, game)
									}
									isFirst = true
								}
							}
						}
					}
				})
			}
		}
	}

	var _exPlayers []model.Women
	model.Connect.Find(&_exPlayers)

	exPlayers := map[string]model.Women{}
	for _, player := range _exPlayers {
		exPlayers[player.Code] = player
	}

	for _, item := range games {
		item.Player1 = exPlayers[item.PlayerCode1].ID
		item.Player1Name = exPlayers[item.PlayerCode1].Name
		item.Player2 = exPlayers[item.PlayerCode2].ID
		item.Player2Name = exPlayers[item.PlayerCode2].Name
		oddAvgMy1, oddAvgMy2, cnt1, cnt2 := calcGame(item)
		if cnt1 == 10 && cnt2 == 10 {
			item.OddAvgMy1 = oddAvgMy1
			item.OddAvgMy2 = oddAvgMy2
			model.Connect.Save(&item)
		} else {
			fmt.Printf("Не хватает игры для расчёта: cnt1 %d, cnt2 %d\n", cnt1, cnt2)
		}
	}
	os.Exit(1)

}
func parserGamesWomenDay(year int, month int, day int) (games []model.WomenGame) {

	var _month string
	if month > 9 {
		_month = fmt.Sprintf("%d", month)
	} else {
		_month = fmt.Sprintf("0%d", month)
	}

	var _day string
	if day > 9 {
		_day = fmt.Sprintf("%d", day)
	} else {
		_day = fmt.Sprintf("0%d", day)
	}
	dateCur, _ := time.Parse(time.RFC3339, fmt.Sprintf("%d-%s-%sT00:00:01Z", year, _month, _day))

	games = []model.WomenGame{}
	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	client := &http.Client{Transport: tr}
	url := fmt.Sprintf("https://www.tennisexplorer.com/results/?type=wta-single&year=%d&month=%s&day=%d", year, _month, day)
	//url = "https://www.tennisexplorer.com/results/?type=wta-single&year=2020&month=02&day=16"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/44.0.2403.157 Safari/537.36")
	res, err := client.Do(req)

	if err == nil {
		defer res.Body.Close()
		if res.StatusCode == 200 {
			doc, err2 := goquery.NewDocumentFromReader(res.Body)
			if err2 == nil {
				var id int
				var matchURL string
				var game model.WomenGame

				scores := map[int][]int{
					0: []int{},
					1: []int{},
				}
				doc.Find("table.result > tbody > tr:not(.head)").Each(func(i int, s *goquery.Selection) {

					s.Find("td:last-child > a").Each(func(i int, q *goquery.Selection) {
						matchURL, _ = q.Attr("href")
						fmt.Sscanf(matchURL, "/match-detail/?id=%d", &id)
					})

					tmpScore := []int{}
					s.Find("td.score").Each(func(j int, q *goquery.Selection) {
						var val int
						var err error
						if val, err = strconv.Atoi(string(strings.Trim(q.Text(), " "))); err == nil {
							tmpScore = append(tmpScore, val)
						}
					})

					href, _ := s.Find("td.t-name > a").Attr("href")
					chunks := strings.Split(href, "/")
					if len(chunks) == 4 {
						if i%2 == 0 {
							game = model.WomenGame{
								DateEvent: dateCur,
								ID:        id,
								URL:       matchURL,
							}
							game.PlayerCode1 = chunks[2]
							scores[0] = tmpScore
						} else {

							scores[1] = tmpScore
							game.PlayerCode2 = chunks[2]

							for i := range scores[0] {
								if scores[0][i] >= 10 {
									scores[0][i] = 1
									scores[1][i] = 0
								}
								game.Scores = game.Scores + fmt.Sprintf("%d:%d;", scores[0][i], scores[1][i])
							}
							if game.Scores != "" {
								games = append(games, game)
							}
							/*if id == 1873433 {
								fmt.Printf("%+v", game.Scores)
								os.Exit(1)
							}*/
							scores = map[int][]int{
								0: []int{},
								1: []int{},
							}
						}
					}
				})
			}
		}
	}

	return
}

func GameWomenYear(year int) {

	var _exPlayers []model.Women
	model.Connect.Find(&_exPlayers)

	exPlayers := map[string]model.Women{}
	for _, player := range _exPlayers {
		exPlayers[player.Code] = player
	}

	var _exGames []model.WomenGame
	model.Connect.Find(&_exPlayers)

	exGames := map[int]model.WomenGame{}
	for _, item := range _exGames {
		exGames[item.ID] = item
	}
	var womens []model.Women
	model.Connect.Find(&womens)
	for _, player := range womens {

		games := parserGamesWomenYear(year, player.Tennisexplorer)
		for _, item := range games {
			item.Player1 = exPlayers[item.PlayerCode1].ID
			item.Player2 = exPlayers[item.PlayerCode2].ID
			model.Connect.Save(&item)
			fmt.Printf("%+v\n", item)
		}

	}
}

func parserGamesWomenYear(year int, url string) (games []model.WomenGame) {

	games = []model.WomenGame{}

	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	client := &http.Client{Transport: tr}
	_url := fmt.Sprintf("%s?annual=%d", url, year)
	req, _ := http.NewRequest("GET", _url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/44.0.2403.157 Safari/537.36")
	res, err := client.Do(req)

	if err == nil {
		defer res.Body.Close()
		if res.StatusCode == 200 {
			doc, err2 := goquery.NewDocumentFromReader(res.Body)
			if err2 == nil {
				selector := fmt.Sprintf("div#matches-%d-1-data > table.balance > tbody > tr", year)
				doc.Find(selector).Each(func(i int, s *goquery.Selection) {
					date := s.Find("td.first").Text()
					if date != "" {
						dates := strings.Split(date, ".")
						str := fmt.Sprintf("%d-%s-%s", year, dates[1], dates[0])
						t1, e := time.Parse("2006-01-02", str)
						if e != nil {
							fmt.Printf("don't parse date %s\n", str)
						}

						var player1Url string
						var player2Url string
						s.Find("td.t-name > a").Each(func(i int, s *goquery.Selection) {
							if i == 0 {
								player1Url, _ = s.Attr("href")
							}
							if i == 1 {
								player2Url, _ = s.Attr("href")
							}
						})

						score := s.Find("td.tl > a")
						matchURL, _ := score.Attr("href")
						var id int
						fmt.Sscanf(matchURL, "/match-detail/?id=%d", &id)

						var scoreResult []string

						r := regexp.MustCompile(`(\<sup\>\w+\<\/sup\>)`)
						_html, _ := score.Html()
						_html = r.ReplaceAllString(_html, "")
						scores := strings.Split(_html, ",")

						for _, set := range scores {
							_games := strings.Split(strings.Trim(set, " "), "-")

							tmpScore := []int{}
							for _, _game := range _games {
								var val int
								var err error
								if val, err = strconv.Atoi(string(strings.Trim(_game, " "))); err == nil {
									tmpScore = append(tmpScore, val)
								}
							}
							if len(tmpScore) > 1 {
								if tmpScore[0] >= 10 {
									tmpScore[0] = 1
									tmpScore[1] = 0
								}
								if tmpScore[1] >= 10 {
									tmpScore[0] = 0
									tmpScore[1] = 1
								}
								scoreResult = append(scoreResult, fmt.Sprintf("%d:%d", tmpScore[0], tmpScore[1]))
							}
						}
						chunks1 := strings.Split(player1Url, "/")
						chunks2 := strings.Split(player2Url, "/")

						games = append(games, model.WomenGame{
							DateEvent:   t1,
							PlayerCode1: chunks1[2],
							PlayerCode2: chunks2[2],
							URL:         matchURL,
							Scores:      strings.Join(scoreResult, ";"),
							ID:          id,
						})
					}
				})
			}
		} else {
			//fmt.Printf("%+v\n", res)
		}
	} else {
		fmt.Printf("%+v\n", err)
	}

	return
}
