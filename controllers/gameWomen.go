package controller

import (
	"fmt"
	"net/http"
	"os"
	"parser/model"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func GamesWomen() {

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

	var year = time.Now().Year()
	var womens []model.Women
	model.Connect.Find(&womens)
	for _, player := range womens {

		games := parserGamesWomen(year, player.Tennisexplorer)
		for _, item := range games {
			item.Player1 = exPlayers[item.PlayerCode1].ID
			item.Player2 = exPlayers[item.PlayerCode2].ID
			model.Connect.Save(&item)
			fmt.Printf("%+v\n", item)
		}

	}
}

func GetGameWomenDay(year int, month int, day int) (games []model.WomenGame) {

	games = []model.WomenGame{}

	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	client := &http.Client{Transport: tr}
	req, _ := http.NewRequest("GET", fmt.Sprintf("https://www.tennisexplorer.com/results/?type=wta-single&year=%d&month=%d&day=%d", year, month, day), nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/44.0.2403.157 Safari/537.36")
	res, err := client.Do(req)

	if err == nil {
		defer res.Body.Close()
		if res.StatusCode == 200 {
			doc, err2 := goquery.NewDocumentFromReader(res.Body)
			if err2 == nil {
				game := model.WomenGame{}
				doc.Find("td > a:last-child").Each(func(i int, s *goquery.Selection) {
					matchURL, _ := s.Attr("href")
					game.URL = matchURL

					var id int
					fmt.Sscanf(matchURL, "/match-detail/?id=%d", &id)
					game.ID = id
				})

				scores := map[int][]int{
					0: []int{},
					1: []int{},
				}
				doc.Find("table.result > tbody > tr:not(.head)").Each(func(i int, s *goquery.Selection) {

					tmpScore := []int{}

					s.Find("td.score").Each(func(j int, q *goquery.Selection) {
						val, _ := strconv.Atoi(string(strings.Trim(q.Text(), "&nbsp;")))
						tmpScore = append(tmpScore, val)
					})

					href, _ := s.Find("td.t-name > a").Attr("href")
					chunks := strings.Split(href, "/")
					fmt.Println(len(chunks))
					if len(chunks) == 4 {
						if i%2 == 0 {
							game = model.WomenGame{}
							game.PlayerCode1 = chunks[2]
							scores[0] = tmpScore
						} else {

							scores[1] = tmpScore
							game.PlayerCode2 = chunks[2]
							for i := range scores[0] {
								if scores[0][i] != 0 && scores[1][i] != 0 {
									game.Scores = game.Scores + fmt.Sprintf("%d:%d;", scores[0][i], scores[1][i])
								}
							}
							games = append(games, game)

							scores = map[int][]int{
								0: []int{},
								1: []int{},
							}
						}
					}

				})
				fmt.Printf("%+v\n", games)
			}
		}
		os.Exit(1)
	}
	return
}
func parserGamesWomen(year int, url string) (games []model.WomenGame) {

	games = []model.WomenGame{}

	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	client := &http.Client{Transport: tr}
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s?annual=%d", url, year), nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/44.0.2403.157 Safari/537.36")
	res, err := client.Do(req)

	if err == nil {
		defer res.Body.Close()
		if res.StatusCode == 200 {
			doc, err2 := goquery.NewDocumentFromReader(res.Body)
			if err2 == nil {
				doc.Find("table.balance > tbody > tr").Each(func(i int, s *goquery.Selection) {
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

						var scoreResult []string
						scores := strings.Split(score.Text(), ",")
						for _, set := range scores {
							_games := strings.Split(strings.Trim(set, " "), "-")
							scoreResult = append(scoreResult, string(_games[0][0])+string(_games[1][0]))
						}
						chunks1 := strings.Split(player1Url, "/")
						chunks2 := strings.Split(player2Url, "/")
						var id int
						fmt.Sscanf(matchURL, "/match-detail/?id=%d", &id)
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
			fmt.Printf("%+v\n", res)
		}
	} else {
		fmt.Printf("%+v\n", err)
	}

	return
}
