package workers

import (
	"time"
)

type WorkStatus string

type WorkStatusReport struct {
	Type      string
	ID        string
	Msg       WorkRequest
	Status    WorkStatus
	StartedOn time.Time
	EndedOn   time.Time
	Messages  []string
	Result    any
	Error     error
}
