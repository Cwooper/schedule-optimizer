package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"time"

	"schedule-optimizer/internal/generator"
	"schedule-optimizer/internal/models"
	"schedule-optimizer/internal/utils"
)

// ----------------------------- SERVER BELOW -----------------------------

func init() {
	// Set initial GC percentage
	debug.SetGCPercent(50)

	// Set soft memory limit
	debug.SetMemoryLimit(2 << 30) // 2GB
}

func main() {
	port := getPort()

	// Serve static files from the frontend directory
	fs := http.FileServer(http.Dir("build"))

	http.HandleFunc("/schedule-optimizer/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			handleScheduleOptimizer(w, r)
		} else {
			// For GET and other methods, serve static files
			if r.URL.Path == "/schedule-optimizer/" {
				http.ServeFile(w, r, "build/index.html")
			} else {
				http.StripPrefix("/schedule-optimizer/", fs).ServeHTTP(w, r)
			}
		}
	})

	log.Printf("Server starting on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// Handles POST requests to /schedule-optimizer
func handleScheduleOptimizer(w http.ResponseWriter, r *http.Request) {
	var request models.RawRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Printf("Error decoding JSON: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Create a context with a 1-second timeout
	ctx, cancel := context.WithTimeout(r.Context(), utils.SERVER_TIMEOUT_SECS*time.Second)
	defer cancel()

	// Create a channel to receive the response
	respChan := make(chan *models.Response)

	// Run GenerateResponse in a goroutine
	go func() {
		respChan <- generator.GenerateResponse(request)
	}()

	// Wait for either the response or a timeout
	select {
	case resp := <-respChan:
		// If we get a response within the timeout, send it

		// Limit the number of schedules sent to the user.
		if len(resp.Schedules) > utils.MAX_OUTPUT_SCHEDULES {
			errString := fmt.Sprintf("There were %d schedules generated, limiting the output to %d schedules. ",
				len(resp.Schedules), utils.MAX_OUTPUT_SCHEDULES)
			errString += "Please narrow your selection criteria."
			resp.Schedules = resp.Schedules[:utils.MAX_OUTPUT_SCHEDULES]
			resp.Warnings = append(resp.Warnings, errString)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	case <-ctx.Done():
		// If we timeout, send an error response
		errorResp := models.Response{
			Errors: []string{"Too many schedules were generated, please narrow your selection criteria."},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusRequestTimeout)
		json.NewEncoder(w).Encode(errorResp)
	}
}

// Get the port to listen on
func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // default port if not specified
	}
	return port
}

// ----------------------- SCHEDULE TESTING BELOW -----------------------------

// Tester function for generating schedules via terminal
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
