package calendar

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var termDateRegex = regexp.MustCompile(`(?i)(Fall|Winter|Spring|Summer):?\s*(\d{1,2}/\d{1,2}/\d{2,4})\s*-\s*(\d{1,2}/\d{1,2}/\d{2,4})`)

// Quarter codes matching jobs.Quarter* constants.
// Duplicated here to avoid an import cycle (jobs imports calendar via setup.go).
const (
	quarterWinter = 10
	quarterSpring = 20
	quarterSummer = 30
	quarterFall   = 40
)

// quarterFromName converts a quarter name (e.g. "Winter", "fall") to its numeric code.
func quarterFromName(name string) int {
	switch strings.ToLower(name) {
	case "winter":
		return quarterWinter
	case "spring":
		return quarterSpring
	case "summer":
		return quarterSummer
	case "fall":
		return quarterFall
	default:
		return 0
	}
}

// parseTermDates extracts term start/end dates from the registrar term-dates page.
func parseTermDates(html []byte) ([]TermDates, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("parse term dates HTML: %w", err)
	}

	var results []TermDates

	doc.Find(".field--name-body").Each(func(_ int, s *goquery.Selection) {
		text := s.Text()
		matches := termDateRegex.FindAllStringSubmatch(text, -1)
		for _, match := range matches {
			quarter := match[1]
			startStr := match[2]
			endStr := match[3]

			startDate, err := parseFlexDate(startStr)
			if err != nil {
				continue
			}
			endDate, err := parseFlexDate(endStr)
			if err != nil {
				continue
			}

			termCode := buildTermCode(quarter, startDate.Year())
			if termCode == "" {
				continue
			}

			results = append(results, TermDates{
				TermCode:  termCode,
				StartDate: startDate,
				EndDate:   endDate,
			})
		}
	})

	if len(results) == 0 {
		return nil, fmt.Errorf("no term dates found in HTML")
	}

	return results, nil
}

// parseFlexDate parses dates in M/D/YYYY or M/D/YY format.
func parseFlexDate(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	for _, layout := range []string{"1/2/2006", "1/2/06"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unparseable date: %s", s)
}

// buildTermCode converts a quarter name and year to a term code like "202610".
func buildTermCode(quarter string, year int) string {
	q := quarterFromName(quarter)
	if q == 0 {
		return ""
	}
	return fmt.Sprintf("%d%02d", year, q)
}

// descriptionToTermCode converts "Winter 2026" to "202610".
func descriptionToTermCode(desc string) string {
	parts := strings.Fields(strings.TrimSpace(desc))
	if len(parts) != 2 {
		return ""
	}
	year, err := strconv.Atoi(parts[1])
	if err != nil {
		return ""
	}
	return buildTermCode(parts[0], year)
}
