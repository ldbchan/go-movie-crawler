package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const TrimSet = " \n\t"

type Host struct {
	host string
}

type Movie struct {
	Name     string
	Desc     string
	Info     string
	Url      string
	ImgUrl   string
	VideoUrl string
	Length   string
	Date     string
}

func save(data []byte, filename string) error {
	return ioutil.WriteFile(filename, data, 0600)
}

func getMovie(url string) Movie {
	doc, err := goquery.NewDocument(url)
	if err != nil {
		fmt.Println("err:", err)
	}

	movie := Movie{}

	// get movie's title
	name := doc.Find(".filmTitle").Text()
	name = strings.Trim(name, TrimSet)
	movie.Name = name

	// get movie's info page url, poster image url, trailer video url
	movie.Url = url
	movie.ImgUrl, _ = doc.Find("#filmTagBlock a.Poster > img").Attr("src")
	movie.VideoUrl, _ = doc.Find("div.video_view > iframe.image.featured").Attr("src")

	// get movie's short description
	desc := doc.Find("#filmTagBlock > span:nth-child(3)").Clone().Children().Remove().End().Text()
	desc = strings.Trim(desc, TrimSet)
	movie.Desc = desc

	// get movie's information
	info := doc.Find("div.video_view").Next().Next().Next().Text()
	info = strings.Replace(info, "劇情簡介", "", -1)
	info = strings.Trim(info, TrimSet)
	movie.Info = info

	doc.Find("#filmTagBlock ul.runtime > li").Each(func(i int, s *goquery.Selection) {
		text := s.Text()

		if strings.Contains(text, "片長") {
			// get movie's length
			length := strings.Replace(text, "分", "", -1)
			length = strings.Replace(length, "片長：", "", -1)
			movie.Length = length
		} else if strings.Contains(text, "上映日期") {
			// get movie's release date
			date := strings.Replace(text, "上映日期：", "", -1)
			movie.Date = date
		}
	})
	// fmt.Println("movie:", movie)
	return movie
}

func saveMoviesFromLinksToFile(links []string, filename string) {
	movies := []Movie{}
	for _, link := range links {
		movie := getMovie(link)
		movies = append(movies, movie)
	}

	// format json data
	jsonData, _ := json.MarshalIndent(movies, "", "  ")

	err := save(jsonData, filename)
	if err != nil {
		fmt.Println("write err:", err)
	} else {
		fmt.Println("saved to", filename)
	}
}

func (h *Host) crawl(findLinkFormat string, url string, filename string) {
	fmt.Printf("start crawling from %v...\n", h.host+url)

	links := []string{}

	doc, err := goquery.NewDocument(h.host + url)
	if err != nil {
		fmt.Println("err:", err)
	}

	doc.Find(findLinkFormat).Each(func(i int, s *goquery.Selection) {
		link, _ := s.Attr("href")
		link = h.host + link

		links = append(links, link)
	})

	saveMoviesFromLinksToFile(links, filename)
}

func main() {
	host := Host{"http://www.atmovies.com.tw"}

	// 近期上映
	host.crawl(".filmNextListAll li a", "/movie/next/0", "coming.json")

	// 首輪
	host.crawl(".filmListAll li a", "/movie/now/0", "first-round.json")

	// 二輪
	host.crawl(".filmListAll li a", "/movie/now2/0", "second-round.json")

	// 本週新片
	host.crawl("article.box.post > div.filmTitle > a", "/movie/new", "new-this-week.json")

	// 新片快報
	host.crawl(".filmNext2ListAll li a", "/movie/next2", "future.json")
}
