package main

import "schedule-optimizer/internal/gpa"

func main() {
	err := gpa.Test()
	if err != nil {
		panic(err)
	}
}