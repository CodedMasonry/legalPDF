package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/PuerkitoBio/goquery"
	"github.com/cenkalti/backoff/v5"
	"github.com/charmbracelet/lipgloss"
)

var (
	tableText = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("34")).Render("Table    ")
	pageText  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("32")).Render("LeafTable")
)

// Either a table or a page, but not both
type crawledItem struct {
	title string
	table *crawledTable
	page  *crawledPage
}

// Represents a crawled table
type crawledTable struct {
	descendants []*crawledItem
}

// Represents a crawled page
type crawledPage struct {
	effective         string
	latestLegislation string
	lines             []string
	footer            string
}

func fetchDocument(url *url.URL) (*goquery.Document, error) {
	fetch := func() (*goquery.Document, error) {
		// An example request that may fail.
		resp, err := http.Get(url.String())
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		// In case on non-retriable error, return Permanent error to stop retrying.
		// For this HTTP example, client errors are non-retriable.
		if resp.StatusCode == 400 {
			return nil, backoff.Permanent(errors.New("bad request"))
		}

		// If we are being rate limited, return a RetryAfter to specify how long to wait.
		// This will also reset the backoff policy.
		if resp.StatusCode == 429 {
			seconds, err := strconv.ParseInt(resp.Header.Get("Retry-After"), 10, 64)
			if err == nil {
				return nil, backoff.RetryAfter(int(seconds))
			}
		}

		// Return successful response.
		return goquery.NewDocumentFromReader(resp.Body)
	}

	result, err := backoff.Retry(context.TODO(), fetch, backoff.WithBackOff(backoff.NewExponentialBackOff()))
	if err != nil {
		return nil, err
	}

	return result, nil
}

// crawls a URL and returns a crawl item
func crawlURL(url *url.URL) (*crawledItem, error) {
	doc, err := fetchDocument(url)
	if err != nil {
		return nil, err
	}

	// Check whether it's a table or a table containing page info
	// Checking whether it can be expanded reduces the number of requests to get same data
	isPageTable := doc.Find("#expand-all-button").Length() > 0
	if !isPageTable {
		// Is a Table
		fmt.Printf("%v %v\n", tableText, url.Path)

		title, table, err := parseTable(url, doc)
		if err != nil {
			return nil, err
		}

		return &crawledItem{
			title: title,

			table: table,
		}, nil
	} else {
		// Is a page
		fmt.Printf("%v %v\n", pageText, url.Path)

		title, pageTable := parsePageTable(doc)
		return &crawledItem{
			title: title,

			table: pageTable,
		}, nil
	}
}

func parseTable(url *url.URL, doc *goquery.Document) (title string, table *crawledTable, err error) {
	title = doc.Find("h1").First().Text()

	var descendants []*crawledItem

	// Select only links pointing to next depth, not references
	doc.Find("td.name-cell a:not([target=\"_blank\"]):not([class])").Each(func(i int, s *goquery.Selection) {
		// Gets a relative URL from the link
		href, err := url.Parse(s.AttrOr("href", ""))
		if err != nil {
			log.Fatalf("Failed to parse href, %v", err)
		}

		// Converts relative to absolute
		absoluteHref := url.ResolveReference(href)

		// Crawl the URL and add it to crawled items
		value, err := crawlURL(absoluteHref)
		if err != nil {
			return
		}
		descendants = append(descendants, value)
	})

	table = &crawledTable{
		descendants: descendants,
	}
	return
}

func parsePageTable(doc *goquery.Document) (title string, table *crawledTable) {
	title = doc.Find("h1").First().Text()
	table = &crawledTable{}

	// Get rows and create crawled Page
	doc.Find("td.name-cell").Each(func(i int, s *goquery.Selection) {
		title, value := parsePage(s)

		table.descendants = append(table.descendants, &crawledItem{
			title: title,
			page:  value,
		})
	})

	return
}

func parsePage(doc *goquery.Selection) (title string, page *crawledPage) {
	title = doc.Find(".content-head-text a").First().Text()
	page = &crawledPage{}

	// Sets effective & Legislation
	info := doc.Find(".laws-section-info-module:not(.no-print) .value")
	page.effective = info.First().Text()
	page.latestLegislation = info.Last().Text()

	// Sets body
	var lines []string
	doc.Find(".laws-body span p").Each(func(i int, s *goquery.Selection) {
		lines = append(lines, s.Text())
	})
	page.lines = lines

	// Sets last updated
	page.footer = doc.Find(".laws-notice p").First().Text()
	return
}
