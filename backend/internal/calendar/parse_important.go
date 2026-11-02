package calendar

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"schedule-optimizer/internal/termcode"
)

// parseTermSelect reads the term dropdown on the important dates page, which
// maps each term to the Drupal node its AJAX view filters on.
func parseTermSelect(html []byte) ([]termOption, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("parse term select HTML: %w", err)
	}
	var options []termOption
	doc.Find("select[name='field_quarter_target_id'] option").Each(func(_ int, s *goquery.Selection) {
		nodeID, _ := s.Attr("value")
		code := termcode.FromDescription(s.Text())
		if nodeID != "" && code != "" {
			options = append(options, termOption{nodeID: nodeID, termCode: code})
		}
	})
	if len(options) == 0 {
		return nil, errors.New("no term options found in select element")
	}
	return options, nil
}

type drupalAjaxCommand struct {
	Command string          `json:"command"`
	Method  string          `json:"method"`
	Data    json.RawMessage `json:"data"`
}

// phaseOneRegex matches the row that opens registration ("Phase I - ..."),
// not Phase II or III.
var phaseOneRegex = regexp.MustCompile(`(?i)^phase\s+(?:i|1)\b`)

// parseImportantDates reads one term's Drupal AJAX response: a list of
// commands, one of which replaces the view with HTML holding a table per
// category. The earliest Phase I row is flagged as the registration opening.
func parseImportantDates(body []byte, code string) ([]importantDate, error) {
	year, quarter, err := termcode.Parse(code)
	if err != nil {
		return nil, err
	}
	e := estimates[quarter]
	anchor := date(year, e.startMonth, e.startDay)

	var commands []drupalAjaxCommand
	if err := json.Unmarshal(body, &commands); err != nil {
		return nil, fmt.Errorf("parse drupal ajax JSON: %w", err)
	}
	var html string
	for _, cmd := range commands {
		if cmd.Command == "insert" && cmd.Method == "replaceWith" &&
			json.Unmarshal(cmd.Data, &html) == nil && html != "" {
			break
		}
	}
	if html == "" {
		return nil, errors.New("no view HTML in AJAX response")
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("parse AJAX HTML: %w", err)
	}

	var dates []importantDate
	opening := -1
	doc.Find("table").Each(func(_ int, table *goquery.Selection) {
		cat := categorizeCaption(table.Find("caption").Text())
		table.Find("tbody tr").Each(func(_ int, tr *goquery.Selection) {
			cells := tr.Find("td")
			description := cleanText(cells.Last().Text())
			start, end, ok := parseDateRange(cleanText(cells.First().Text()), anchor)
			if !ok || description == "" {
				return
			}
			if cat == categoryRegistration && phaseOneRegex.MatchString(description) &&
				(opening < 0 || start.Before(dates[opening].date)) {
				opening = len(dates)
			}
			dates = append(dates, importantDate{date: start, end: end, description: description, category: cat})
		})
	})
	if len(dates) == 0 {
		return nil, errors.New("no dated rows in AJAX HTML")
	}
	if opening >= 0 {
		dates[opening].opensRegistration = true
	}
	return dates, nil
}

// categorizeCaption maps the view's table captions ("Start of Quarter",
// "Registration", "Last Day Deadlines", "End of Quarter", "Holidays", and
// summer's per-session deadline tables) to categories.
func categorizeCaption(caption string) category {
	lower := strings.ToLower(caption)
	switch {
	case strings.Contains(lower, "holiday"):
		return categoryHoliday
	case strings.Contains(lower, "registration"):
		return categoryRegistration
	case strings.Contains(lower, "of quarter"):
		return categoryAcademic
	default:
		return categoryDeadline
	}
}

// Matches "Tue, Nov 3"; the weekday is optional.
var drupalDateRegex = regexp.MustCompile(`^(?:[A-Za-z]{3},?\s+)?([A-Za-z]+)\s+(\d{1,2})$`)

// parseDateRange parses "Tue, Mar 31" or "Tue, Feb 24 - Mon, Mar 9". end is
// zero for a single date.
func parseDateRange(text string, anchor time.Time) (start, end time.Time, ok bool) {
	first, second, isRange := strings.Cut(text, " - ")
	if start, ok = parseMonthDay(first, anchor); !ok {
		return time.Time{}, time.Time{}, false
	}
	if !isRange {
		return start, time.Time{}, true
	}
	if end, ok = parseMonthDay(second, anchor); !ok {
		return time.Time{}, time.Time{}, false
	}
	if end.Before(start) {
		end = end.AddDate(1, 0, 0)
	}
	return start, end, true
}

// parseMonthDay resolves a yearless date to the year that puts it nearest
// anchor, the term's estimated start. The page omits years, and a term's
// dates straddle New Year: Winter 2027 registration opens Nov 3, 2026,
// while its June Commencement deadline is May 16, 2027. Every row observed
// falls within five months of the start, well inside the six that keep
// nearest unambiguous.
func parseMonthDay(text string, anchor time.Time) (time.Time, bool) {
	m := drupalDateRegex.FindStringSubmatch(text)
	if m == nil {
		return time.Time{}, false
	}
	month, ok := parseMonth(m[1])
	day, err := strconv.Atoi(m[2])
	if !ok || err != nil {
		return time.Time{}, false
	}

	var best time.Time
	for _, year := range []int{anchor.Year() - 1, anchor.Year(), anchor.Year() + 1} {
		t := date(year, month, day)
		if t.Month() != month {
			continue // Feb 29 outside a leap year
		}
		if best.IsZero() || absDuration(t.Sub(anchor)) < absDuration(best.Sub(anchor)) {
			best = t
		}
	}
	return best, !best.IsZero()
}

func parseMonth(s string) (time.Month, bool) {
	for _, layout := range []string{"Jan", "January"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t.Month(), true
		}
	}
	return 0, false
}

func absDuration(d time.Duration) time.Duration {
	if d < 0 {
		return -d
	}
	return d
}
