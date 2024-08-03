package main

import (
	"fmt"

	"schedule-optimizer/internal/generator"
	"schedule-optimizer/internal/models"
)

func main() {
	// Gets the request as json post request, use this for now
	request := models.RawRequest{
		Courses: []string{"CSCI 301", "CSCI 247", "CSCI 241"},
		Forced:  []string{},
		Term:    "202440",
		Min:     2,
		Max:     3,
	}

	resp := generator.GenerateResponse(request)
	fmt.Printf("Response: %v\n", resp)
	// send the response back as a json
}

// func main() {
// 	courses := []string{"CSCI 330", "CSCI 345", "CSCI 367", "CSCI 305",
// 		"CSCI 145", "MATH 204", "CSCI 141", "CSCI 241", "CSCI 247"}
// 	forced := []string{}
// 	req := models.RawRequest{
// 		Courses: courses,
// 		Forced:  forced,
// 		Term:    "202440",
// 		Min:     6,
// 		Max:     7,
// 	}
// 	g := generator.NewGenerator()
// 	response := g.GenerateResponse(req)
// 	fmt.Printf("\nErrors: \n\n")
// 	for _, err := range response.Errors {
// 		fmt.Printf("%v\n", err)
// 	}

// 	fmt.Printf("\nWarnings: \n\n")
// 	for _, warning := range response.Warnings {
// 		fmt.Printf("%v\n", warning)
// 	}

// 	fmt.Printf("\nSchedules (%v): \n\n", len(response.Schedules))
// 	for _, schedule := range response.Schedules[:3] {
// 		fmt.Println("Courses: ")
// 		for _, course := range schedule.Courses {
// 			fmt.Printf("%v: ", course.Subject)
// 			for _, session := range course.Sessions {
// 				fmt.Printf("%v ", session)
// 			}
// 			fmt.Println()
// 		}
// 		fmt.Printf("Weights: %v\n", schedule.Weights)
// 		fmt.Printf("Score: %v\n", schedule.Score)
// 		fmt.Println()
// 	}

// 	bytesResponse, _ := json.Marshal(response)
// 	size := uintptr(cap(bytesResponse))*unsafe.Sizeof(bytesResponse) + unsafe.Sizeof(response)
// 	fmt.Printf("Size of response: %v KB\n", size/1000)
// }
