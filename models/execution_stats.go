package models

import "time"

type ExecutionStats struct {
	Processed int
	Skipped   int
	StartTime time.Time
}
