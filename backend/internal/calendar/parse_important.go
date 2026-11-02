package calendar

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"schedule-optimizer/internal/termcode"
)

const termSelect = "select[name='field_quarter_target_id']"

// parseTermSelect reads the term dropdown on the important dates page, which
// maps each term to the node ID the page filters on.
func parseTermSelect(html []byte) ([]termOption, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("parse term select HTML: %w", err)
	}
	var options []termOption
	doc.Find(termSelect + " option").Each(func(_ int, s *goquery.Selection) {
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

// phaseOneRegex matches the row that opens registration ("Phase I - ..."),
// not Phase II or III.
var phaseOneRegex = regexp.MustCompile(`(?i)^phase\s+(?:i|1)\b`)

// parseImportantDates reads the important dates page filtered to one term:
// a table per category, each row a date or range and a description. The
// earliest Phase I row is flagged as the registration opening.
//
// It fails on any row it cannot read, since storing the rest would delete
// that row (Phase I, say) from the term. It also fails unless the page's
// selected term is code, so a page that ignored the filter cannot file one
// term's dates under another.
func parseImportantDates(html []byte, code string) ([]importantDate, error) {
	year, quarter, err := termcode.Parse(code)
	if err != nil {
		return nil, err
	}
	e := estimates[quarter]
	anchor := date(year, e.startMonth, e.startDay)

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("parse important dates HTML: %w", err)
	}
	selected := cleanText(doc.Find(termSelect + " option[selected]").First().Text())
	if termcode.FromDescription(selected) != code {
		return nil, fmt.Errorf("page shows term %q", selected)
	}

	var dates []importantDate
	var errs []error
	opening := -1
	doc.Find("table").Each(func(_ int, table *goquery.Selection) {
		cat := categorizeCaption(table.Find("caption").Text())
		table.Find("tbody tr").Each(func(_ int, tr *goquery.Selection) {
			cells := tr.Find("td")
			when, description := cleanText(cells.First().Text()), cleanText(cells.Last().Text())
			if when == "" && description == "" {
				return
			}
			start, end, ok := parseDateRange(when, anchor)
			if !ok || description == "" {
				errs = append(errs, fmt.Errorf("unreadable row %q %q", when, description))
				return
			}
			if cat == categoryRegistration && phaseOneRegex.MatchString(description) &&
				(opening < 0 || start.Before(dates[opening].date)) {
				opening = len(dates)
			}
			dates = append(dates, importantDate{date: start, end: end, description: description, category: cat})
		})
	})
	if err := errors.Join(errs...); err != nil {
		return nil, err
	}
	if len(dates) == 0 {
		return nil, errors.New("no dated rows in HTML")
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

// parseDateRange parses "Tue, Mar 31" or "Tue, Feb 24 - Mon, Mar 9", with a
// hyphen or en dash. end is zero for a single date.
func parseDateRange(text string, anchor time.Time) (start, end time.Time, ok bool) {
	first, second, isRange := strings.Cut(strings.ReplaceAll(text, "\u2013", "-"), "-")
	first, second = strings.TrimSpace(first), strings.TrimSpace(second)
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
