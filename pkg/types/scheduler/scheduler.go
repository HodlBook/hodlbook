package scheduler

import "time"

type Scheduler interface {
	Start() error
	Stop()
}

const (
	IntervalMinute = 1 * time.Minute
	IntervalDaily  = 24 * time.Hour
)
