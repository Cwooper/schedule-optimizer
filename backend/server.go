package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"sync"
	"time"

	"github.com/robfig/cron"

	"github.com/cwooper/schedule-optimizer/internal/generator"
	"github.com/cwooper/schedule-optimizer/internal/models"
	"github.com/cwooper/schedule-optimizer/internal/scraper"
	"github.com/cwooper/schedule-optimizer/internal/search"
	"github.com/cwooper/schedule-optimizer/internal/utils"
)

// ----------------------------- SERVER BELOW -----------------------------

var (
	isUpdating  bool
	updateMutex sync.RWMutex
)

func init() {
	// Set initial GC percentage
	debug.SetGCPercent(50)
	// Set soft memory limit
	debug.SetMemoryLimit(2 << 30) // 2GB
}

func main() {
	// Optionally run update-courses with update-courses
	// This is NOT Safe is a server is running.
	if len(os.Args) > 1 && os.Args[1] == "update-courses" {
		UpdateCoursesHandler()
		return
	}

	updateSchedule := fmt.Sprintf("0 %s %s * * *", utils.UPDATE_MIN, utils.UPDATE_HOUR)
	c := cron.New()
	c.AddFunc(updateSchedule, func() {
		log.Println("Starting scheduled course update.")
		UpdateCoursesHandler()
	})
	c.Start()

	port := getPort()
	fs := http.FileServer(http.Dir("../build"))

	http.HandleFunc("/schedule-optimizer/subjects", func(w http.ResponseWriter, r *http.Request) {
		content, err := os.ReadFile("../data/subjects.txt")
		if err != nil {
			log.Printf("Error reading subjects file: %v", err)
			http.Error(w, "Error reading subjects file", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		w.Write(content)
	})

	http.HandleFunc("/schedule-optimizer/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			handleScheduleOptimizer(w, r)
		} else {
			if r.URL.Path == "/schedule-optimizer/" {
				http.ServeFile(w, r, "../build/index.html")
			} else {
				http.StripPrefix("/schedule-optimizer/", fs).ServeHTTP(w, r)
			}
		}
	})

	log.Printf("Server starting on port %s\n", port)
	nextRun := c.Entries()[0].Next
	log.Printf("Next scheduled run: %s", nextRun.Format(time.RFC3339))
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// Handles POST requests to /schedule-optimizer
func handleScheduleOptimizer(w http.ResponseWriter, r *http.Request) {
	// Checks if the server is currently updating data and sends error if so
	updateMutex.RLock()
	updating := isUpdating
	updateMutex.RUnlock()

	if updating {
		errorResp := models.Response{
			Errors: []string{"Please wait one minute... Updating course data."},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(errorResp)
		return
	}

	// If we're not updating, continue with the request and generate schedules'
	var request models.RawRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		log.Printf("Error decoding JSON: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if request.SearchTerm != "" { // Fuzzy search
		Search(w, r, request)
	} else if len(request.Courses) > 0 { // Schedule Generator
		ScheduleGenerator(w, r, request)
	} else { // Unknown
		errorResp := models.Response{
			Errors: []string{"Error parsing your request (not search or schedule generator?)"},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResp)
	}
}

// Fuzzy search for courses
func Search(w http.ResponseWriter, r *http.Request, request models.RawRequest) {
	resp := search.SearchCourses(request.SearchTerm, request.Term)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Generate schedules
func ScheduleGenerator(w http.ResponseWriter, r *http.Request, request models.RawRequest) {
	// Context with an n second timeout for extra-large queries
	ctx, cancel := context.WithTimeout(r.Context(), utils.SERVER_TIMEOUT_SECS*time.Second)
	defer cancel()

	respChan := make(chan *models.Response)
	go func() {
		respChan <- generator.GenerateResponse(request)
	}()

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
		port = "48920" // default port if not specified
	}
	return port
}

// Attempts to update the course data by webscraping (if necessary)
// updateMutex assures thread safety and allows for continued uptime.
func UpdateCoursesHandler() {
	updateMutex.Lock()
	checkUpdating := isUpdating
	updateMutex.Unlock()
	// Check for updating twice before finished previous update
	if checkUpdating {
		log.Println("Tried to update while already updating.")
		os.Exit(1)
	}
	// We are currently updating
	updateMutex.Lock()
	isUpdating = true
	updateMutex.Unlock()

	defer func() {
		updateMutex.Lock()
		isUpdating = false
		updateMutex.Unlock()
	}()

	err := scraper.UpdateCourses()
	if err != nil {
		log.Printf("Error updating courses: %v", err)
	} else {
		log.Println("Courses updated successfully")
	}
}
