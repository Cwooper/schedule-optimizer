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

var (
	// Finals week heading: "December 7-11", with an en dash on the live page.
	finalsWeekRegex = regexp.MustCompile(`([A-Za-z]+)\s+(\d{1,2})\s*[-\x{2013}]`)

	// Row header: "Between 8:00 and 8:29 AM".
	classRangeRegex = regexp.MustCompile(`(\d{1,2}:\d{2})\s+and\s+(\d{1,2}:\d{2})\s*(AM|PM)`)

	// Cell: "Monday 10:30 - 12:30 PM".
	examSlotRegex = regexp.MustCompile(`(?i)(Monday|Tuesday|Wednesday|Thursday|Friday)\s*(\d{1,2}:\d{2})\s*[-\x{2013}]\s*(\d{1,2}:\d{2})\s*(AM|PM)`)
)

// parseFinals reads the finals page: one <wwu-switcher-item label="Fall 2026">
// per term, each holding an <h2> naming finals week and a table whose rows are
// class start-time ranges and whose columns split classes by whether they meet
// Tuesday or Thursday. It returns the mappings of every term that parsed in
// full, and the failures confined to one term.
func parseFinals(html []byte) ([]finalMapping, []termError, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(html))
	if err != nil {
		return nil, nil, fmt.Errorf("parse finals HTML: %w", err)
	}

	var results []finalMapping
	var failed []termError
	terms := 0
	doc.Find("wwu-switcher-item").Each(func(_ int, item *goquery.Selection) {
		label, _ := item.Attr("label")
		code := termcode.FromDescription(label)
		if code == "" {
			return
		}
		terms++
		mappings, err := parseFinalsTerm(item, code)
		if err != nil {
			failed = append(failed, termError{code: code, err: err})
			return
		}
		results = append(results, mappings...)
	})
	if terms == 0 {
		return nil, nil, errors.New("no finals terms found in HTML")
	}
	return results, failed, nil
}

// parseFinalsTerm fails on any row or cell it cannot read, since storing the
// rest would delete that slot's stored mapping.
func parseFinalsTerm(item *goquery.Selection, code string) ([]finalMapping, error) {
	year, _, _ := termcode.Parse(code)
	week := finalsWeekRegex.FindStringSubmatch(cleanText(item.Find("h2").First().Text()))
	if week == nil {
		return nil, errors.New("no finals week heading")
	}
	monday, err := time.ParseInLocation("January 2 2006", fmt.Sprintf("%s %s %d", week[1], week[2], year), Pacific)
	if err != nil {
		return nil, fmt.Errorf("finals week heading: %w", err)
	}
	// Exam cells name only a weekday, so dates count from the week's Monday.
	for monday.Weekday() != time.Monday {
		monday = monday.AddDate(0, 0, -1)
	}

	table := item.Find("table").First()
	// The header names which column holds classes meeting neither Tuesday
	// nor Thursday; read it rather than trusting column order.
	var hasTuTh []bool
	table.Find("thead th").Each(func(i int, th *goquery.Selection) {
		if i > 0 {
			hasTuTh = append(hasTuTh, !strings.Contains(strings.ToLower(th.Text()), "neither"))
		}
	})

	var results []finalMapping
	var errs []error
	table.Find("tbody tr").Each(func(_ int, tr *goquery.Selection) {
		header := cleanText(tr.Find("th").First().Text())
		r := classRangeRegex.FindStringSubmatch(header)
		if r == nil {
			errs = append(errs, fmt.Errorf("unreadable class time %q", header))
			return
		}
		cells := tr.Find("td")
		if cells.Length() != len(hasTuTh) {
			errs = append(errs, fmt.Errorf("row %q has %d exam cells, header has %d", header, cells.Length(), len(hasTuTh)))
			return
		}
		rangeStart, rangeEnd := timeRange(r[1], r[2], r[3])
		cells.Each(func(col int, td *goquery.Selection) {
			text := cleanText(td.Text())
			slot := examSlotRegex.FindStringSubmatch(text)
			if slot == nil {
				errs = append(errs, fmt.Errorf("unreadable exam slot %q", text))
				return
			}
			examStart, examEnd := timeRange(slot[2], slot[3], slot[4])
			results = append(results, finalMapping{
				termCode:   code,
				rangeStart: rangeStart,
				rangeEnd:   rangeEnd,
				hasTuTh:    hasTuTh[col],
				examDate:   monday.AddDate(0, 0, weekdayOffset(slot[1])),
				examStart:  examStart,
				examEnd:    examEnd,
			})
		})
	})
	if err := errors.Join(errs...); err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, errors.New("no exam slots in table")
	}
	return results, nil
}

func weekdayOffset(day string) int {
	for i := range 5 {
		if strings.EqualFold(time.Weekday(i+1).String(), day) {
			return i
		}
	}
	return 0
}

// timeRange converts "10:30", "12:30", "PM" to "1030", "1230". The page writes
// one AM/PM marker for both ends, and it belongs to the end: a start that
// would land after the end under the same marker is in the morning.
func timeRange(start, end, marker string) (string, string) {
	endMin := minutesOfDay(end, marker)
	startMin := minutesOfDay(start, marker)
	if startMin > endMin {
		startMin -= 12 * 60
	}
	return hhmm(startMin), hhmm(endMin)
}

func minutesOfDay(clock, marker string) int {
	h, m, _ := strings.Cut(clock, ":")
	hour, _ := strconv.Atoi(h)
	minute, _ := strconv.Atoi(m)
	hour %= 12
	if strings.EqualFold(marker, "PM") {
		hour += 12
	}
	return hour*60 + minute
}

func hhmm(minutes int) string {
	return fmt.Sprintf("%02d%02d", minutes/60, minutes%60)
}
