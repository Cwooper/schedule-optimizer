package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/cwooper/schedule-optimizer/internal/api"
	"github.com/cwooper/schedule-optimizer/internal/cache"
	"github.com/cwooper/schedule-optimizer/internal/models"
)

func main() {
	fmt.Println("Course Finder - Interactive Search")
	fmt.Println("----------------------------------")

	// Initialize API client
	client, err := api.NewClient()
	if err != nil {
		fmt.Printf("Warning: Could not initialize API client: %v\n", err)
		fmt.Println("Will use only cached data.")
	}

	// Initialize cache
	courseManager := cache.GetInstance()

	reader := bufio.NewReader(os.Stdin)
	for {
		// Get term
		fmt.Print("\nEnter term (e.g., 202520 for Spring 2025, or 'q' to quit): ")
		term, err := reader.ReadString('\n')
		term = strings.TrimSpace(term)
		if err != nil || term == "q" {
			break
		}

		// Get subject
		fmt.Print("Enter subject (e.g., CSCI, or press Enter for all): ")
		subject, err := reader.ReadString('\n')
		subject = strings.TrimSpace(subject)
		if err != nil || subject == "q" {
			break
		}

		// Get course number
		fmt.Print("Enter course number (e.g., 141, or press Enter for all): ")
		courseNum, err := reader.ReadString('\n')
		courseNum = strings.TrimSpace(courseNum)
		if err != nil  || subject == "q" {
			break
		}

		// Try to get courses from cache first
		var courses []models.Course
		courses, err = courseManager.GetCourseList(term)
		if err != nil {
			if client != nil {
				// If cache fails and we have a client, try API
				fmt.Println("Fetching from API...")
				if subject == "" {
					subject = "%"
				}
				if courseNum == "" {
					courseNum = "%"
				}
				courses, err = client.GetCourses(term, subject, courseNum)
				if err != nil {
					fmt.Printf("Error fetching courses: %v\n", err)
					continue
				}
			} else {
				fmt.Printf("Error: No cached data for term %s and API client not available\n", term)
				continue
			}
		}

		if len(courses) == 0 {
			fmt.Println("No courses found matching criteria.")
			continue
		}

		// Display results using tabwriter for nice formatting
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "\nCourse\tTitle\tInstructor\tGPA\tTime\tLocation")
		fmt.Fprintln(w, "------\t-----\t----------\t---\t----\t--------")

		for _, course := range courses {
			var timeStr, locStr string
			if len(course.Sessions) > 0 && course.Sessions[0].Days != "" {
				session := course.Sessions[0]
				timeStr = fmt.Sprintf("%s %04d-%04d", session.Days, session.StartTime, session.EndTime)
				locStr = session.Location
			} else if course.Sessions[0].IsTimeTBD {
				timeStr = "TBA"
				locStr = "TBA"
			} else if course.Sessions[0].IsAsync {
				timeStr = "Online"
				locStr = "Online"
			} else {
				timeStr = "Unknown"
				locStr = "Unknown"
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%f\t%s\t%s\n",
				course.Subject,
				course.Title,
				course.Instructor,
				course.GPA,
				timeStr,
				locStr,
			)

			// If course has multiple sessions, show additional times indented
			for i := 1; i < len(course.Sessions); i++ {
				session := course.Sessions[i]
				timeStr = fmt.Sprintf("%s %04d-%04d", session.Days, session.StartTime, session.EndTime)
				fmt.Fprintf(w, "\t\t\t\t%s\t%s\n",
					timeStr,
					session.Location,
				)
			}
		}
		w.Flush()
		fmt.Printf("\nFound %d courses.\n", len(courses))
	}

	fmt.Println("\nGoodbye!")
}
