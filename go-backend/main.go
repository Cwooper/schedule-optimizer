package main

import (
	"schedule-optimizer/internal/scraper"
)

func main() {
	err := scraper.UpdateCourses()
	if err != nil {
		panic(err)
	}
}