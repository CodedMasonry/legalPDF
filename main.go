package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/PuerkitoBio/goquery"
)

type crawledItem struct {
	table *table
	page  *page
}

type table struct {
	title       string
	descendants []*crawledItem
}

type page struct {
	title     string
	effective string
	lines     []string
}

func main() {
	url, _ := url.Parse("https://codes.ohio.gov/ohio-revised-code")

	parsed := crawlURL(url)
	fmt.Println(parsed)
}

func crawlURL(url *url.URL) *crawledItem {
	fmt.Printf("Crawling: %v\n", url.String())
	res, err := http.Get(url.String())
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		log.Fatalf("Non-200 error code: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	rows := doc.Find(".name-cell a")
	if rows.Length() > 0 {
		// Is a Table
		return &crawledItem{
			table: parseTable(url, doc),
		}
	} else {
		// Is a page
		return &crawledItem{
			page: parsePage(doc),
		}
	}
}

func parseTable(url *url.URL, doc *goquery.Document) *table {
	title := doc.Find("h1").First().Text()

	var descendants []*crawledItem
	doc.Find(".name-cell a").Each(func(i int, s *goquery.Selection) {
		// Gets a relative URL from the link
		href, err := url.Parse(s.AttrOr("href", ""))
		if err != nil {
			log.Fatalf("Failed to parse href, %v", err)
		}

		// Converts relative to absolute
		absoluteHref := url.ResolveReference(href)

		// Crawl the URL and add it to crawled items
		value := crawlURL(absoluteHref)
		descendants = append(descendants, value)
	})

	return &table{
		title:       title,
		descendants: descendants,
	}
}

func parsePage(doc *goquery.Document) *page {
	title := doc.Find("h1").First().Text()
	effective := doc.Find("laws-section-info-module value").First().Text()
	var lines []string

	doc.Find("laws-body p").Each(func(i int, s *goquery.Selection) {
		lines = append(lines, s.Text())
	})

	doc.Find(".name-cell a")

	return &page{
		title:     title,
		effective: effective,
		lines:     lines,
	}
}
