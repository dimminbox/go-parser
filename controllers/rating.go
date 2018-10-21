package controller

import (
	"net/http"
	"parser/model"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const RATING_URL = "https://www.atpworldtour.com/en/rankings/singles"

func GetRating() {

	beginDate := time.Now().AddDate(0, -1, 0)
	//duration := time.Now().Sub(beginDate)

	var _exPlayers []model.Player
	model.Connect.Find(&_exPlayers)
	exPlayers := map[string]int{}
	for _, player := range _exPlayers {
		exPlayers[player.Code] = player.ID
	}

	for beginDate.Unix() < time.Now().Unix() {

		var dateUpdate string
		beginDate = beginDate.AddDate(-0, 0, 1)

		if beginDate.Weekday().String() == "Monday" {

			dateUpdate = beginDate.Format("2006-01-02")

			var _exRatings []model.Rating
			model.Connect.Where("dateUpdate = ?", dateUpdate).Find(&_exRatings)
			exRatings := map[string]int{}
			for _, _rating := range _exRatings {

				uniqKey := _rating.DateUpdate + strconv.Itoa(_rating.Player)
				exRatings[uniqKey] = _rating.ID
			}

			ratings := parseRating(dateUpdate)
			for _, rating := range ratings {

				if _, ok := exPlayers[rating.Code]; ok {
					rating.Player = exPlayers[rating.Code]

					curKey := rating.DateUpdate + strconv.Itoa(rating.Player)
					if existRating, isExist := exRatings[curKey]; isExist {
						rating.ID = existRating
					}
					model.Connect.Save(&rating)
				}
			}

			time.Sleep(100 * time.Millisecond)
		}
	}

}

func parseRating(dateRating string) (ratings []model.Rating) {

	rRating, _ := regexp.Compile(`(\d+)`)
	r := strings.NewReplacer(",", "", ")", ",", " ", "")

	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	client := &http.Client{Transport: tr}

	url := RATING_URL + "?rankDate=" + dateRating + "&countryCode=all&rankRange=0-2000"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/44.0.2403.157 Safari/537.36")
	res, err := client.Do(req)
	if err == nil {
		defer res.Body.Close()
		if res.StatusCode == 200 {

			doc, err2 := goquery.NewDocumentFromReader(res.Body)
			if err2 == nil {
				doc.Find("div.table-rankings-wrapper table.mega-table tbody tr").Each(func(i int, s *goquery.Selection) {
					rating := model.Rating{DateUpdate: dateRating}

					s.Find("td.rank-cell").Each(func(i int, s *goquery.Selection) {

						numbers := rRating.FindStringSubmatch(s.Text())

						if len(numbers) == 2 {
							rating.Rating, _ = strconv.Atoi(numbers[1])
						} else {
							return
						}

					})

					s.Find("td.points-cell a").Each(func(i int, s *goquery.Selection) {
						rating.Points, _ = strconv.Atoi(r.Replace(strings.TrimSpace(s.Text())))

					})

					s.Find("td.player-cell a").Each(func(i int, s *goquery.Selection) {

						href, _ := s.Attr("href")
						hrefArr := strings.Split(href, "/")
						if len(hrefArr) > 3 {
							rating.Code = hrefArr[len(hrefArr)-2]
						}
					})
					ratings = append(ratings, rating)
				})

			}
		}
	}

	return
}
