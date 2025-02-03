package main

import (
	"net/url"
	"time"

	"github.com/go-pdf/fpdf"
)

var (
	font = "Times"
)

type PDF struct {
	inner *fpdf.Fpdf
}

// Creates a PDF and sets a front page
func createPDF(url *url.URL, item *crawledItem) *PDF {
	// Init
	pdf := fpdf.New("P", "mm", "A4", "")

	// Settings
	pdf.SetTopMargin(30)

	// Add front page
	pdf.AddPage()

	// Header
	pdf.SetFont(font, "B", 30)
	pdf.Cell(0, 0, item.title)
	pdf.Ln(10)

	// Subtitles
	pdf.SetFont(font, "B", 12)
	pdf.Cell(20, 0, "Parsed: ")
	pdf.SetFont(font, "I", 12)
	pdf.Cell(0, 0, time.Now().Format("January _2, 2006"))
	pdf.Ln(6)
	pdf.SetFont(font, "B", 12)
	pdf.Cell(20, 0, "Source: ")
	pdf.SetFont(font, "I", 12)
	pdf.Cell(0, 0, url.Host)
	pdf.Ln(12)

	// Render the crawled data
	for _, row := range item.table.descendants {
		renderItem(pdf, row, 0)
	}

	return &PDF{
		inner: pdf,
	}
}

func renderItem(pdf *fpdf.Fpdf, item *crawledItem, depth int) {
	// Set Title
	if depth == 0 {
		pdf.SetLeftMargin(10)
		pdf.SetFont(font, "B", 18)
	} else if depth == 1 {
		pdf.SetLeftMargin(14)
		pdf.SetFont(font, "B", 16)
	} else {
		pdf.SetLeftMargin(18)
		pdf.SetFont(font, "BI", 14)
	}
	pdf.MultiCell(0, 6, item.title, "", "", false)

	if item.table != nil {
		// Render children
		pdf.Ln(6)
		for _, row := range item.table.descendants {
			renderItem(pdf, row, depth+1)
		}
	} else {
		// Render Page

		// Sets law information
		pdf.Ln(6)
		pdf.SetFont(font, "BI", 10)
		pdf.Cell(18, 0, "Effective: ")
		pdf.SetFont(font, "I", 10)
		pdf.Cell(30, 0, item.page.effective)
		pdf.SetFont(font, "BI", 10)
		pdf.Cell(30, 0, "Latest Legislation: ")
		pdf.SetFont(font, "I", 10)
		pdf.Cell(30, 0, item.page.latestLegislation)
		pdf.Ln(6)

		// Seperator
		pdf.Line(pdf.GetX(), pdf.GetY(), pdf.GetX()+180, pdf.GetY())
		pdf.Ln(2)

		// Renders actual text
		pdf.SetFont(font, "", 10)
		for _, line := range item.page.lines {
			pdf.MultiCell(180, 6, line, "", "", false)
			pdf.Ln(4)
		}

		// Renders last updated
		if item.page.footer != "" {
			pdf.SetFont(font, "I", 8)
			pdf.Cell(0, 0, item.page.footer)
			pdf.Ln(6)
		}

		// Seperator
		pdf.Line(pdf.GetX(), pdf.GetY(), pdf.GetX()+180, pdf.GetY())
		pdf.Ln(6)
	}
}

func (p *PDF) writeFile(path string) error {
	return p.inner.OutputFileAndClose(path)
}
