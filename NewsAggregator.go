package main

import (
	"encoding/xml"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
)

var wg sync.WaitGroup

type SitemapIndex struct {
	Locations []string `xml:"sitemap>loc"`
}

type News struct {
	Titles    []string `xml:"url>news>title"`
	Locations []string `xml:"url>loc"`
}

type NewsMap struct {
	Location string
}

type NewsAggPage struct {
	Title string
	News  map[string]NewsMap
}

func newsRoutine(c chan News, Location string) {
	defer wg.Done()
	var n News
	link := strings.TrimSpace(Location)
	resp, _ := http.Get(link)
	bytes, _ := ioutil.ReadAll(resp.Body)
	xml.Unmarshal(bytes, &n)
	resp.Body.Close()

	c <- n

}

func newsAggHandler(w http.ResponseWriter, r *http.Request) {
	var s SitemapIndex
	news_map := make(map[string]NewsMap)

	resp, _ := http.Get("https://www.washingtonpost.com/news-sitemaps/index.xml")
	bytes, _ := ioutil.ReadAll(resp.Body)
	xml.Unmarshal(bytes, &s)
	resp.Body.Close()

	queue := make(chan News, 30)

	for _, Location := range s.Locations {
		wg.Add(1)
		go newsRoutine(queue, Location)
	}

	wg.Wait()
	close(queue)

	for elem := range queue {
		for idx, _ := range elem.Locations {
			news_map[elem.Titles[idx]] = NewsMap{elem.Locations[idx]}
		}
	}

	p := NewsAggPage{Title: "News Aggregator", News: news_map}
	t, _ := template.ParseFiles("NewsApp.gohtml")
	fmt.Println(t.Execute(w, p))
}

func main() {
	port := os.Getenv("PORT")
	http.HandleFunc("/", newsAggHandler)
	http.ListenAndServe(":"+port, nil)
}

// func main() {
// 	// port := os.Getenv("PORT")
// 	http.HandleFunc("/", newsAggHandler)
// 	http.ListenAndServe(":1000", nil)
// }
