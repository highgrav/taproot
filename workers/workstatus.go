package workers

import (
	"time"
)

type WorkStatus string

type WorkStatusReport struct {
	ID        string
	Status    WorkStatus
	StartedOn time.Time
	EndedOn   time.Time
	Messages  []string
	Result    any
	Error     error
}
