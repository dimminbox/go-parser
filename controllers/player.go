package controller

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"parser/model"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const ATP_PLAYERS_LIST_URL = "https://www.atpworldtour.com/en/rankings/singles?countryCode=all&rankRange="
const MYSCORE_PLAYERS_LIST_URL = "https://www.myscore.ru/tennis/rankings/atp/"

func GetWorldPlayers(count int) (Players []model.Player) {

	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	client := &http.Client{Transport: tr}
	res, err := client.Get(ATP_PLAYERS_LIST_URL + "0-" + strconv.Itoa(count))
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

	doc.Find("div.table-rankings-wrapper table.mega-table tbody tr ").Each(func(i int, s *goquery.Selection) {

		title := strings.TrimSpace(s.Find("td.player-cell").Text())

		r1 := strings.NewReplacer("\t", "", "\n", "")
		move, _ := strconv.Atoi(strings.TrimSpace(r1.Replace(s.Find("td.move-cell div.move-text").Text())))

		moveUp := s.Find("td.move-cell div.move-up").Length()
		moveDown := s.Find("td.move-cell div.move-down").Length()

		if moveUp == 0 && moveDown == 1 {
			move = move * -1
		}
		href, _ := s.Find("td.player-cell a ").Attr("href")
		href = model.ATP_PLAYERS_URL + href
		r := strings.NewReplacer("\n", "", " ", "", ",", "", "\t", "")
		points, _ := strconv.Atoi(r.Replace(s.Find("td.points-cell").Text()))
		Players = append(Players, model.Player{Name: title, Tennisworld: href, Points: points, MoveRank: move, Sex: 1})
	})

	return
}

func GetMyscorePlayers(Players []model.Player) (Players1 []model.Player) {

	Players1 = Players

	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	client := &http.Client{Transport: tr}

	res, err := client.Get(MYSCORE_PLAYERS_LIST_URL)
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

	doc.Find("div#ranking-table table tbody tr.rank-row").Each(func(i int, s *goquery.Selection) {

		if len(Players1) >= (i + 1) {
			href, _ := s.Find("td.rank-column-player a ").Attr("href")
			href = model.MYSCORE_PLAYERS_URL + href
			Players1[i].Rank = i + 1
			Players1[i].Myscore = href
		}
	})

	return
}

func GetPlayers(count int) (players map[string]model.Player) {

	players = map[string]model.Player{}

	players1 := GetWorldPlayers(count)
	players1 = GetMyscorePlayers(players1)

	ch := make(chan model.Player)
	for _, player := range players1 {
		go parsePlayer(player, ch)
		time.Sleep(100 * time.Millisecond)
	}

	for i := 0; i < len(players1); i++ {
		data := <-ch
		players[data.Tennisworld] = data
	}
	//fmt.Println(players)
	return
}

func parsePlayer(player model.Player, ch chan model.Player) {

	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	client := &http.Client{Transport: tr}

	req, _ := http.NewRequest("GET", player.Tennisworld, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/44.0.2403.157 Safari/537.36")

	res, err1 := client.Do(req)
	if err1 == nil {
		defer res.Body.Close()
		if res.StatusCode == 200 {
			doc, err2 := goquery.NewDocumentFromReader(res.Body)
			if err2 == nil {

				doc.Find("div.player-flag-code").Each(func(i int, s *goquery.Selection) {
					player.Coutry = s.Text()
				})

				doc.Find("div.wrap").Each(func(i int, s *goquery.Selection) {
					isAge := s.Find("div.table-big-label").Text()
					if isAge == "Age" {
						age := s.Find("div.table-big-value").Text()
						age = strings.Replace(strings.Replace(age, " ", "", -1), "\n", "", -1)
						player.Age = age
					}
				})

				doc.Find("div.wrap").Each(func(i int, s *goquery.Selection) {
					isPro := s.Find("div.table-big-label").Text()
					if isPro == "Turned Pro" {
						pro := s.Find("div.table-big-value").Text()
						pro = strings.Replace(strings.Replace(pro, " ", "", -1), "\n", "", -1)
						player.Pro = pro
					}
				})

				doc.Find("div.wrap").Each(func(i int, s *goquery.Selection) {
					isWeight := s.Find("div.table-big-label").Text()
					if isWeight == "Weight" {
						r := strings.NewReplacer("(", "", ")", "", "k", "", "g", "")
						weight := r.Replace(s.Find("span.table-weight-kg-wrapper").Text())
						player.Weight = weight
					}
				})

				doc.Find("div.wrap").Each(func(i int, s *goquery.Selection) {
					isHeight := s.Find("div.table-big-label").Text()
					if isHeight == "Height" {
						r := strings.NewReplacer("(", "", ")", "", "c", "", "m", "")
						height := r.Replace(s.Find("span.table-height-cm-wrapper").Text())
						player.Height = height
					}
				})

				doc.Find("div.wrap").Each(func(i int, s *goquery.Selection) {
					isBirth := s.Find("div.table-label").Text()
					r := strings.NewReplacer("\t", "", "\n", "", " ", "")
					isBirth = r.Replace(isBirth)
					if isBirth == "Birthplace" {
						player.Birthplace = r.Replace(s.Find("div.table-value").Text())
					}
				})

				doc.Find("div.wrap").Each(func(i int, s *goquery.Selection) {
					isHand := s.Find("div.table-label").Text()
					r := strings.NewReplacer("\t", "", "\n", "", " ", "")
					isHand = r.Replace(isHand)
					if isHand == "Plays" {
						r1 := strings.NewReplacer("\t", "", "\n", "")
						player.Hand = strings.TrimSpace(r1.Replace(s.Find("div.table-value").Text()))
					}
				})

				doc.Find("div.wrap").Each(func(i int, s *goquery.Selection) {
					isCoach := s.Find("div.table-label").Text()
					r := strings.NewReplacer("\t", "", "\n", "", " ", "")
					isCoach = r.Replace(isCoach)
					if isCoach == "Coach" {
						r1 := strings.NewReplacer("\t", "", "\n", "")
						player.Coach = strings.TrimSpace(r1.Replace(s.Find("div.table-value").Text()))
					}
				})

				doc.Find("div.player-profile-hero-image img").Each(func(i int, s *goquery.Selection) {

					configuration := model.Init()
					player.Photo = strings.Replace(player.Name, " ", "_", -1)
					filePath := configuration.Resources + "/images/player/" + player.Photo
					f, err := os.Create(filePath)
					defer f.Close()
					if err != nil {
						log.Fatal(err)
					} else {

						photo, _ := s.Attr("src")
						photoRes, err := http.Get(model.ATP_PLAYERS_URL + photo)
						defer photoRes.Body.Close()
						if err == nil {
							body, _ := ioutil.ReadAll(photoRes.Body)
							f.Write(body)
						}

					}
				})

				doc.Find("table#playersStatsTable tbody tr td").Each(func(i int, s *goquery.Selection) {

					switch i {
					case 3:
						winLose := s.Find("div.stat-value").Text()
						winLose = strings.TrimSpace(winLose)
						winLoseArr := strings.Split(winLose, "-")
						player.WinCY, _ = strconv.Atoi(winLoseArr[0])
						player.LoseCY, _ = strconv.Atoi(winLoseArr[1])

					case 8:
						winLose := s.Find("div.stat-value").Text()
						winLose = strings.TrimSpace(winLose)
						winLoseArr := strings.Split(winLose, "-")
						player.Win, _ = strconv.Atoi(winLoseArr[0])
						player.Lose, _ = strconv.Atoi(winLoseArr[1])

					case 4:
						titles, _ := strconv.Atoi(strings.TrimSpace(s.Find("div.stat-value").Text()))
						player.Titles = titles

					case 9:
						titles, _ := strconv.Atoi(strings.TrimSpace(s.Find("div.stat-value").Text()))
						player.TitlesCY = titles

					case 7:
						rank, _ := strconv.Atoi(strings.TrimSpace(s.Find("div.stat-value").Text()))
						player.HighRank = rank

					case 5:
						r := strings.NewReplacer("\t", "", "\n", "", " ", "", ",", "", "$", "")
						money, _ := strconv.Atoi(strings.TrimSpace(r.Replace(s.Find("div.stat-value").Text())))
						player.PrizeMoneyCY = money

					case 10:
						r := strings.NewReplacer("\t", "", "\n", "", " ", "", ",", "", "$", "")
						money, _ := strconv.Atoi(strings.TrimSpace(r.Replace(s.Find("div.stat-value").Text())))
						player.PrizeMoney = money

					}
				})

			}

		}
	}

	ch <- player

}

func Players(count int) {

	players := GetPlayers(count)

	var _exPlayers []model.Player
	model.Connect.Find(&_exPlayers)

	exPlayers := map[string]model.Player{}
	for _, player := range _exPlayers {
		exPlayers[player.Tennisworld] = player
	}

	for _, player := range players {
		if exPlayer, ok := exPlayers[player.Tennisworld]; ok {
			player.ID = exPlayer.ID
		}
		model.Connect.Save(&player)
		fmt.Printf("%d - %s \n", player.ID, player.Name)
	}

}
