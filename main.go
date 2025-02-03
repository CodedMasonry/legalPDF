package main

import (
	"log"
	"net/url"
)

func main() {
	url, _ := url.Parse("https://codes.ohio.gov/ohio-revised-code")

	data, err := crawlURL(url)
	if err != nil {
		log.Fatal(err)
	}
	pdf := createPDF(url, data)

	if err := pdf.writeFile("test.pdf"); err != nil {
		log.Fatal(err)
	}
}
