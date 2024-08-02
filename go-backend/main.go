package main

import "schedule-optimizer/internal/gpa"

func main() {
	_, err := gpa.GetGPAData()
	if err != nil {
		panic(err)
	}
}
