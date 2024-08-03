package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"schedule-optimizer/internal/generator"
	"schedule-optimizer/internal/models"
)

func main() {
	port := getPort()
	// Serve static files from the frontend directory
	fs := http.FileServer(http.Dir("frontend"))
	http.Handle("/schedule-optimizer/", http.StripPrefix("/schedule-optimizer/", fs))

	// Handle GET and POST requests for /schedule-optimizer
	http.HandleFunc("/schedule-optimizer", handleScheduleOptimizer)

	log.Printf("Server starting on port %s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

func handleScheduleOptimizer(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		// Serve the index.html file for GET requests
		http.ServeFile(w, r, filepath.Join("frontend", "index.html"))
	case "POST":
		// Handle POST requests for optimization
		var request models.RawRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		resp := generator.GenerateResponse(request)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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
