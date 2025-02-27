package main

import (
	"github.com/cwooper/schedule-optimizer/internal/api"
)

// The entire purpose of this package is to update the protobufs
// without starting the server.

func main() {
	api.UpdateCourses()
}
