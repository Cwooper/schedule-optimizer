package calendar

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"schedule-optimizer/internal/termcode"
)

// Matches "Winter: 1/6/26 - 3/20/2026"; later years use "4/2/2030- 6/14/2030".
var termDateRegex = regexp.MustCompile(`(?i)(Fall|Winter|Spring|Summer):?\s*(\d{1,2}/\d{1,2}/\d{2,4})\s*-\s*(\d{1,2}/\d{1,2}/\d{2,4})`)

func parseTermDates(html []byte) ([]termDates, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("parse term dates HTML: %w", err)
	}

	var results []termDates
	for _, m := range termDateRegex.FindAllStringSubmatch(cleanText(doc.Find(".field--name-body").Text()), -1) {
		start, errStart := parseSlashDate(m[2])
		end, errEnd := parseSlashDate(m[3])
		if err := errors.Join(errStart, errEnd); err != nil {
			return nil, err
		}
		results = append(results, termDates{
			termCode: termcode.FromName(m[1], start.Year()),
			start:    start,
			end:      end,
		})
	}
	if len(results) == 0 {
		return nil, errors.New("no term dates found in HTML")
	}
	return results, nil
}

// parseSlashDate parses M/D/YYYY or M/D/YY; the page mixes both.
func parseSlashDate(s string) (time.Time, error) {
	for _, layout := range []string{"1/2/2006", "1/2/06"} {
		if t, err := time.ParseInLocation(layout, s, Pacific); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unparseable date %q", s)
}

// cleanText collapses whitespace. strings.Fields splits on U+00A0, which the
// regexps' \s does not match; the page puts &nbsp; between some labels and
// dates.
func cleanText(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
