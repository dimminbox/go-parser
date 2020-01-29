package controller

import (
	"fmt"
	"log"
	"net/http"
	"parser/model"
	"time"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const LIMIT = 27
const BASE_URL = "https://www.tennisexplorer.com"
const LIST_URL = "/ranking/wta-women/"

func GetWomens(date string, page int) (Players []model.Women) {

	Players = []model.Women{}

	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	client := &http.Client{Transport: tr}

	url := fmt.Sprintf("%s%s?date=%s&page=%d", BASE_URL, LIST_URL, date, page)

	res, err := client.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	doc.Find("tbody.flags > tr ").Each(func(i int, s *goquery.Selection) {
		item := s.Find("td.t-name > a")
		href, _ := item.Attr("href")

		hrefArr := strings.Split(href, "/")
		var code string
		if len(hrefArr) > 3 {
			code = hrefArr[len(hrefArr)-2]
		} else {
			code = ""
		}

		player := model.Women{
			Tennisexplorer: BASE_URL + href,
			Name:           item.Text(),
			Country:        s.Find("td.tl").Text(),
			Code : code,
		}
		Players = append(Players, player)
	})

	return
}

func Womens(date string) {

	players := []model.Women{}

	for i := 1; i < LIMIT; i++ {
		players = append(players, GetWomens(date, i)...)
	}


	var _exPlayers []model.Women
	model.Connect.Find(&_exPlayers)

	exPlayers := map[string]model.Women{}
	for _, player := range _exPlayers {
		exPlayers[player.Code] = player
	}

	for _, player := range players {
		if exPlayer, ok := exPlayers[player.Code]; ok {
			player.ID = exPlayer.ID
		}
		model.Connect.Save(&player)
		fmt.Printf("%d - %s \n", player.ID, player.Name)
	}

	/*for _, player := range players{
		model.Connect.Save(&player)
	}*/

}
