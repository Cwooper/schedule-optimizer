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
var termDateRegex = regexp.MustCompile(`(?i)(Fall|Winter|Spring|Summer):?\s*(\d{1,2}/\d{1,2}/\d{2,4})\s*[-\x{2013}]\s*(\d{1,2}/\d{1,2}/\d{2,4})`)

// Terms run 8 to 12 weeks. A span outside 4 to 20 is a typo on the
// hand-edited page (a wrong year, most likely), and storing it could mark a
// current term past and stop its scrapes.
const (
	minTermDays = 28
	maxTermDays = 140
)

// parseTermDates returns the terms whose dates parse, and the failures
// confined to one term.
func parseTermDates(html []byte) ([]termDates, []termError, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(html))
	if err != nil {
		return nil, nil, fmt.Errorf("parse term dates HTML: %w", err)
	}

	matches := termDateRegex.FindAllStringSubmatch(cleanText(doc.Find(".field--name-body").Text()), -1)
	if len(matches) == 0 {
		return nil, nil, errors.New("no term dates found in HTML")
	}
	var results []termDates
	var failed []termError
	for _, m := range matches {
		start, errStart := parseSlashDate(m[2])
		end, errEnd := parseSlashDate(m[3])
		if err := errors.Join(errStart, errEnd); err != nil {
			// No reliable year, so no term code to attribute it to.
			failed = append(failed, termError{err: fmt.Errorf("%s: %w", m[1], err)})
			continue
		}
		code := termcode.FromName(m[1], start.Year())
		if days := end.Sub(start).Hours() / 24; days < minTermDays || days > maxTermDays {
			failed = append(failed, termError{code: code, err: fmt.Errorf("%s - %s spans %.0f days", m[2], m[3], days)})
			continue
		}
		results = append(results, termDates{termCode: code, start: start, end: end})
	}
	return results, failed, nil
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
