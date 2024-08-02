package main

import (
	"fmt"
	"schedule-optimizer/internal/generator"
	"schedule-optimizer/internal/models"
)

func main() {
	req := models.RawRequest{
		Courses: []string{"CSCI 301", "CSCI 241", "CSCI 141", "CSCI 247"},
		Forced:  nil,
		Term:    "202440",
		Min:     2,
		Max:     3,
	}
	g := generator.NewGenerator()
	response := g.GenerateResponse(req)
	fmt.Printf("\nErrors: \n\n")
	for _, err := range response.Errors {
		fmt.Printf("%v\n", err)
	}

	fmt.Printf("\nWarnings: \n\n")
	for _, warning := range response.Warnings {
		fmt.Printf("%v\n", warning)
	}

	fmt.Printf("\nSchedules: \n\n")
	for _, schedule := range response.Schedules {
		fmt.Println("Courses: ")
		for _, course := range schedule.Courses {
			fmt.Printf("%v %v\n", course.Subject, course.Sessions[0].Instructor)
		}
		fmt.Printf("Weights: %v\n", schedule.Weights)
		fmt.Printf("Score: %v\n", schedule.Score)
		fmt.Println()
	}
}
