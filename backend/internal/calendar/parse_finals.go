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

var (
	// Matches finals week heading like "March 16–20" or "June 8–12" or "December 7-11"
	finalsWeekRegex = regexp.MustCompile(`(\w+)\s+(\d{1,2})\s*[–-]\s*(\d{1,2})`)

	// Matches time range in row headers: "8:00 and 8:29 AM" or "4:30 and 5:59 PM"
	timeRangeRegex = regexp.MustCompile(`(\d{1,2}:\d{2})\s+and\s+(\d{1,2}:\d{2})\s*(AM|PM)`)

	// Matches exam slot in cells: "Friday\n8:00 - 10:00 AM"
	examSlotRegex = regexp.MustCompile(`(?i)(Monday|Tuesday|Wednesday|Thursday|Friday)\s*(\d{1,2}:\d{2})\s*-\s*(\d{1,2}:\d{2})\s*(AM|PM)`)

	// Matches "Winter 2026" etc.
	finalsTermRegex = regexp.MustCompile(`(?i)(Winter|Spring|Summer|Fall)\s+(\d{4})`)
)

// parseFinals extracts finals schedule mappings from the registrar finals page.
// The page uses <wwu-switcher-item label="Winter 2026"> elements containing
// an <h2> with the finals week date range and a <table> with the schedule.
func parseFinals(html []byte) ([]FinalMapping, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("parse finals HTML: %w", err)
	}

	var results []FinalMapping

	doc.Find("table").Each(func(_ int, table *goquery.Selection) {
		termCode := ""
		var finalsWeekStart time.Time

		// Check for wwu-switcher-item parent with label attribute
		parent := table.Parent()
		for range 5 {
			if label, exists := parent.Attr("label"); exists {
				termMatch := finalsTermRegex.FindStringSubmatch(label)
				if termMatch != nil {
					termCode = descriptionToTermCode(termMatch[1] + " " + termMatch[2])
				}
				break
			}
			parent = parent.Parent()
		}

		if termCode == "" {
			containerText := table.Parent().Text()
			termMatch := finalsTermRegex.FindStringSubmatch(containerText)
			if termMatch != nil {
				termCode = descriptionToTermCode(termMatch[1] + " " + termMatch[2])
			}
		}

		if termCode == "" {
			return
		}

		year, _ := strconv.Atoi(termCode[:4])

		// Find the finals week heading — look for h2 before this table
		table.PrevAll().Each(func(_ int, el *goquery.Selection) {
			if !finalsWeekStart.IsZero() {
				return
			}
			if goquery.NodeName(el) != "h2" {
				return
			}
			weekMatch := finalsWeekRegex.FindStringSubmatch(el.Text())
			if weekMatch == nil {
				return
			}
			dateStr := fmt.Sprintf("%s %s, %d", weekMatch[1], weekMatch[2], year)
			t, err := time.Parse("January 2, 2006", dateStr)
			if err == nil {
				finalsWeekStart = t
			}
		})

		if finalsWeekStart.IsZero() {
			return
		}

		dayToDate := buildFinalsWeekMap(finalsWeekStart)

		table.Find("tbody tr").Each(func(_ int, tr *goquery.Selection) {
			th := tr.Find("th").First()
			rangeMatch := timeRangeRegex.FindStringSubmatch(th.Text())
			if rangeMatch == nil {
				return
			}

			rangeStart := to24h(rangeMatch[1], rangeMatch[3])
			rangeEnd := to24h(rangeMatch[2], rangeMatch[3])

			tr.Find("td").Each(func(colIdx int, td *goquery.Selection) {
				if colIdx > 1 {
					return
				}

				slotMatch := examSlotRegex.FindStringSubmatch(td.Text())
				if slotMatch == nil {
					return
				}

				examDate, ok := dayToDate[strings.ToLower(slotMatch[1])]
				if !ok {
					return
				}

				results = append(results, FinalMapping{
					TermCode:       termCode,
					TimeRangeStart: rangeStart,
					TimeRangeEnd:   rangeEnd,
					HasTuTh:        colIdx == 0,
					ExamDate:       examDate,
					ExamStartTime:  to24h(slotMatch[2], slotMatch[4]),
					ExamEndTime:    to24h(slotMatch[3], slotMatch[4]),
				})
			})
		})
	})

	if len(results) == 0 {
		return nil, fmt.Errorf("no finals mappings found in HTML")
	}

	return results, nil
}

// buildFinalsWeekMap returns a map from lowercase day name to date for a M-F finals week.
func buildFinalsWeekMap(monday time.Time) map[string]time.Time {
	for monday.Weekday() != time.Monday {
		monday = monday.AddDate(0, 0, -1)
	}
	return map[string]time.Time{
		"monday":    monday,
		"tuesday":   monday.AddDate(0, 0, 1),
		"wednesday": monday.AddDate(0, 0, 2),
		"thursday":  monday.AddDate(0, 0, 3),
		"friday":    monday.AddDate(0, 0, 4),
	}
}

// to24h converts "8:00" + "AM" to "0800", "1:00" + "PM" to "1300", etc.
func to24h(timeStr, ampm string) string {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return timeStr
	}
	hour, err := strconv.Atoi(parts[0])
	if err != nil {
		return timeStr
	}
	min, err := strconv.Atoi(parts[1])
	if err != nil {
		return timeStr
	}

	if strings.EqualFold(ampm, "PM") && hour != 12 {
		hour += 12
	}
	if strings.EqualFold(ampm, "AM") && hour == 12 {
		hour = 0
	}

	return fmt.Sprintf("%02d%02d", hour, min)
}
