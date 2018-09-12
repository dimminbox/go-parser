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
)

const GAME_URL = "http://www.atpworldtour.com"

func Games(year int) {

	var tournaments []model.Tournament

	games := []model.Game{}
	model.Connect.Where("year = ?", year).Find(&tournaments)

	tournaments = tournaments[150:151]
	for _, tournament := range tournaments {
		games = append(games, parserGames(tournament.Tennisworld)...)
	}

	ch := make(chan model.Game)

	games = games[0:1]
	//fmt.Printf("%+v\n", games)
	for _, game := range games {
		go parserGame(game, ch)
		time.Sleep(1000 * time.Millisecond)
	}

}

func parserGames(url string) (games []model.Game) {

	games = []model.Game{}

	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	client := &http.Client{Transport: tr}

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/44.0.2403.157 Safari/537.36")
	res, err := client.Do(req)
	if err == nil {
		defer res.Body.Close()
		if res.StatusCode == 200 {

			doc, err2 := goquery.NewDocumentFromReader(res.Body)
			if err2 == nil {

				doc.Find("table.day-table tbody tr").Each(func(i int, s *goquery.Selection) {
					game := model.Game{}
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

			} else {
				log.Println(err1)
			}

		}
	} else {
		log.Println(err)
	}

	fmt.Printf("%#v\n", game)
	ch <- game

}
