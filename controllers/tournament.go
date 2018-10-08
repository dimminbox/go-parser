package controller

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"parser/model"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const TOURNAMENT_ATP_URL = "http://www.atpworldtour.com/-/ajax/Scores/GetTournamentArchiveForYear/"
const TOURNAMENT_URL = "http://www.atpworldtour.com/en/scores/archive/"
const TOURNAMENT_CH_URL = "http://www.atpworldtour.com/en/scores/results-archive?tournamentType=ch&ajax=true"

type Tournir struct {
	Value          string `json:"Value"`
	Key            string `json:"Key"`
	DataAttributes string `json:"-"`
}
type Tournirs []Tournir

func Tournaments(year int) {

	tournaments := GetTournaments(year)

	var _exTournaments []model.Tournament
	model.Connect.Find(&_exTournaments)

	exTournaments := map[string]model.Tournament{}
	for _, _tournament := range _exTournaments {
		exTournaments[_tournament.Tennisworld] = _tournament
	}

	for _, tournament := range tournaments {
		if exTournament, ok := exTournaments[tournament.Tennisworld]; ok {
			tournament.ID = exTournament.ID
		}
		model.Connect.Save(&tournament)
		fmt.Printf("%d - %s \n", tournament.ID, tournament.Name)
	}

}

func GetTournaments(year int) (tournaments map[string]model.Tournament) {

	tournaments = map[string]model.Tournament{}

	tournirs := GetATP(year)
	tournirs = append(tournirs, GetChallenger(year)...)

	ch := make(chan model.Tournament)
	for _, tour := range tournirs {

		tournir := model.Tournament{}
		tournir.Year = year
		tournir.Tennisworld = TOURNAMENT_URL + tour.Key + "/" + tour.Value + "/" + strconv.Itoa(year) + "/results"

		go parserTournament(tournir, ch)
		time.Sleep(100 * time.Millisecond)
	}

	for i := 0; i < len(tournirs); i++ {
		data := <-ch
		tournaments[data.Tennisworld] = data
	}
	return
}

func parserTournament(tournament model.Tournament, ch chan model.Tournament) {

	regexType := regexp.MustCompile(`categorystamps_([\w\s]+)_`)

	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	client := &http.Client{Transport: tr}

	req, _ := http.NewRequest("GET", tournament.Tennisworld, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/44.0.2403.157 Safari/537.36")
	res, err1 := client.Do(req)

	if err1 == nil {
		defer res.Body.Close()
		if res.StatusCode == 200 {
			doc, err2 := goquery.NewDocumentFromReader(res.Body)
			if err2 == nil {

				codeMyscore := ""
				myscoreArr := strings.Split(strings.Replace(strings.ToLower(tournament.Tennisworld), " ", "-", -1), "/")
				if len(myscoreArr) > 6 {
					codeMyscore = myscoreArr[6]
				}

				doc.Find("span.tourney-dates").Each(func(i int, s *goquery.Selection) {
					dates := strings.Split(strings.TrimSpace(s.Text()), "-")

					if len(dates) == 2 {
						tournament.DateStart = strings.Replace(strings.TrimSpace(dates[0]), ".", "-", -1)
						tournament.DateStop = strings.Replace(strings.TrimSpace(dates[1]), ".", "-", -1)
					}
				})

				doc.Find("td.tourney-badge-wrapper > img").Each(func(i int, s *goquery.Selection) {
					image, _ := s.Attr("src")
					imageType := regexType.FindStringSubmatch(image)

					if len(imageType) == 2 {

						switch imageType[1] {
						case "250":
							tournament.Type = "ATP 250"
							tournament.Hospitality = 1
							tournament.Myscore = "https://www.myscore.ru/tennis/atp-singles/" + codeMyscore + "-" + strconv.Itoa(tournament.Year) + "/results/"
						case "500":
							tournament.Type = "ATP 500"
							tournament.Hospitality = 1
							tournament.Myscore = "https://www.myscore.ru/tennis/atp-singles/" + codeMyscore + "-" + strconv.Itoa(tournament.Year) + "/results/"
						case "grandslam":
							tournament.Type = "Grand Slam"
							tournament.Hospitality = 1
							tournament.Myscore = "https://www.myscore.ru/tennis/atp-singles/" + codeMyscore + "-" + strconv.Itoa(tournament.Year) + "/results/"
						case "1000s":
							tournament.Type = "ATP 1000"
							tournament.Hospitality = 1
							tournament.Myscore = "https://www.myscore.ru/tennis/atp-singles/" + codeMyscore + "-" + strconv.Itoa(tournament.Year) + "/results/"
						case "ATP Challenger", "challenger":
							tournament.Type = "ATP Challenger"
							tournament.Myscore = "https://www.myscore.ru/tennis/challenger-men-singles/" + codeMyscore + "-" + strconv.Itoa(tournament.Year) + "/results/"
						default:
							tournament.Hospitality = 1
						}
					}

				})
			}

			doc.Find(".tourney-title").Each(func(i int, s *goquery.Selection) {
				tournament.Name = strings.TrimSpace(s.Text())
			})

			doc.Find("td.tourney-details div.info-area div.item-details a.not-in-system").Each(func(i int, s *goquery.Selection) {

				switch i {
				case 0:
					tournament.Sgl = strings.TrimSpace(s.Text())
				case 1:
					tournament.Dbl = strings.TrimSpace(s.Text())
				}

			})

			doc.Find("div.info-area div.item-details span.item-value").Each(func(i int, s *goquery.Selection) {
				switch i {
				case 2:
					tournament.Surface = strings.TrimSpace(s.Text())
				case 3:
					r := strings.NewReplacer("\t", "", "\n", "", " ", "", ",", "")
					money := strings.TrimSpace(r.Replace(strings.TrimSpace(s.Text())))
					tournament.PrizeMoney = money
				}

			})

		}
	}

	ch <- tournament
}
func GetATP(year int) Tournirs {

	tournirs := &Tournirs{}

	url := TOURNAMENT_ATP_URL + strconv.Itoa(year)

	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/44.0.2403.157 Safari/537.36")
	res, err := client.Do(req)
	if err == nil {
		defer res.Body.Close()
		body, _ := ioutil.ReadAll(res.Body)
		json.Unmarshal(body, tournirs)
	}

	return *tournirs
}

func GetChallenger(year int) (tournirs Tournirs) {

	url := TOURNAMENT_CH_URL + "&year=" + strconv.Itoa(year)

	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/44.0.2403.157 Safari/537.36")
	res, err := client.Do(req)
	if err == nil {
		doc, err1 := goquery.NewDocumentFromReader(res.Body)
		if err1 == nil {

			doc.Find("td.tourney-details > a").Each(func(i int, s *goquery.Selection) {
				href, _ := s.Attr("href")
				hrefArr := strings.Split(href, "/")
				if len(hrefArr) > 5 {
					tournirs = append(tournirs, Tournir{Value: hrefArr[5], Key: hrefArr[4]})
				}
			})

		}
	}

	return
}
