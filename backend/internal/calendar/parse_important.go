package calendar

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// parseDrupalTermSelect extracts the term option mapping from the
// important-dates-deadlines page's <select> element.
func parseDrupalTermSelect(html []byte) ([]DrupalTermOption, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(html))
	if err != nil {
		return nil, fmt.Errorf("parse term select HTML: %w", err)
	}

	var options []DrupalTermOption

	doc.Find("select[name='field_quarter_target_id'] option").Each(func(_ int, s *goquery.Selection) {
		val, exists := s.Attr("value")
		if !exists || val == "" || val == "All" {
			return
		}
		desc := strings.TrimSpace(s.Text())
		termCode := descriptionToTermCode(desc)
		if termCode == "" {
			return
		}
		options = append(options, DrupalTermOption{
			NodeID:      val,
			Description: desc,
			TermCode:    termCode,
		})
	})

	if len(options) == 0 {
		return nil, fmt.Errorf("no term options found in select element")
	}

	return options, nil
}

// drupalAjaxCommand represents one command in the Drupal AJAX response array.
type drupalAjaxCommand struct {
	Command string          `json:"command"`
	Method  string          `json:"method"`
	Data    json.RawMessage `json:"data"`
}

// parseImportantDatesAjax parses the Drupal AJAX response for a term.
// Returns holidays and important dates separately.
func parseImportantDatesAjax(jsonBody []byte, termCode string) ([]Holiday, []ImportantDate, error) {
	var commands []drupalAjaxCommand
	if err := json.Unmarshal(jsonBody, &commands); err != nil {
		return nil, nil, fmt.Errorf("parse drupal ajax JSON: %w", err)
	}

	// Find the insert/replaceWith command containing the HTML
	var htmlData string
	for _, cmd := range commands {
		if cmd.Command == "insert" && cmd.Method == "replaceWith" && len(cmd.Data) > 2 {
			if err := json.Unmarshal(cmd.Data, &htmlData); err != nil {
				continue
			}
			if htmlData != "" {
				break
			}
		}
	}
	if htmlData == "" {
		return nil, nil, fmt.Errorf("no insert command found in AJAX response")
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlData))
	if err != nil {
		return nil, nil, fmt.Errorf("parse AJAX HTML data: %w", err)
	}

	var holidays []Holiday
	var importantDates []ImportantDate

	doc.Find("table").Each(func(_ int, table *goquery.Selection) {
		caption := strings.TrimSpace(table.Find("caption").Text())
		category := categorizeCaption(caption)

		table.Find("tbody tr").Each(func(_ int, tr *goquery.Selection) {
			dateCell := tr.Find("td").First()
			eventCell := tr.Find("td").Last()

			dateText := cleanText(dateCell.Text())
			eventText := cleanText(eventCell.Text())

			if dateText == "" || eventText == "" {
				return
			}

			date := parseDrupalDate(dateText, termCode)
			if date.IsZero() {
				return
			}

			if category == "holiday" {
				holidays = append(holidays, Holiday{
					TermCode:    termCode,
					Date:        date,
					Description: eventText,
				})
			} else {
				importantDates = append(importantDates, ImportantDate{
					TermCode:    termCode,
					Date:        date,
					Description: eventText,
					Category:    category,
				})
			}
		})
	})

	return holidays, importantDates, nil
}

// categorizeCaption maps table captions to categories.
func categorizeCaption(caption string) string {
	lower := strings.ToLower(caption)
	switch {
	case strings.Contains(lower, "holiday"):
		return "holiday"
	case strings.Contains(lower, "registration"):
		return "registration"
	case strings.Contains(lower, "start") || strings.Contains(lower, "end") ||
		strings.Contains(lower, "quarter") || strings.Contains(lower, "final") ||
		strings.Contains(lower, "commencement") || strings.Contains(lower, "break"):
		return "academic"
	default:
		return "deadline"
	}
}

var drupalDateRegex = regexp.MustCompile(`(?i)(?:Mon|Tue|Wed|Thu|Fri|Sat|Sun),?\s+(\w+)\s+(\d{1,2})`)

// parseDrupalDate parses dates like "Tue, Mar 31" or "Tue, Feb 24 - Mon, Mar 9".
// For date ranges, returns the start date only (end dates are not stored).
func parseDrupalDate(text string, termCode string) time.Time {
	if len(termCode) < 4 {
		return time.Time{}
	}
	year, err := strconv.Atoi(termCode[:4])
	if err != nil {
		return time.Time{}
	}

	// Replace non-breaking spaces
	text = strings.ReplaceAll(text, "\u00a0", " ")

	// For ranges like "Tue, Feb 24 - Mon, Mar 9", use the start date
	parts := strings.SplitN(text, " - ", 2)
	return parseSingleDrupalDate(parts[0], year)
}

func parseSingleDrupalDate(text string, year int) time.Time {
	match := drupalDateRegex.FindStringSubmatch(text)
	if match == nil {
		return time.Time{}
	}

	dateStr := fmt.Sprintf("%s %s, %d", match[1], match[2], year)

	for _, layout := range []string{"January 2, 2006", "Jan 2, 2006"} {
		if t, err := time.Parse(layout, dateStr); err == nil {
			return t
		}
	}
	return time.Time{}
}

// cleanText normalizes whitespace and trims a string.
func cleanText(s string) string {
	s = strings.ReplaceAll(s, "\u00a0", " ")
	return strings.Join(strings.Fields(s), " ")
}
