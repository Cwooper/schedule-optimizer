package main

import (
	"fmt"

	"github.com/cwooper/schedule-optimizer/internal/api"
)

func main() {
	// Create a new API client
	client, err := api.NewClient()
	if err != nil {
		panic(err)
	}

	// Initialize the session
	courses, err := client.GetCourses("202520", "GEOL", "212")
	if err != nil {
		panic(err)
	}
	
	// Print the courses
	for _, course := range courses {
		fmt.Println(course)
	}
}
