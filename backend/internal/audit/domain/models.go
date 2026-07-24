package domain

import "time"

type AuditEntry struct {
	ID              string
	EventType       string
	ActorID         string
	ActorEmail      string
	Action          string
	ResourceType    string
	ResourceID      string
	ResourceChanges string
	Metadata        map[string]string
	IPAddress       string
	UserAgent       string
	TraceID         string
	CreatedAt       time.Time
}

type AuditStats struct {
	TotalEntries int32
	ByEventType  map[string]int32
	ByAction     map[string]int32
	UniqueActors int32
	DateFrom     time.Time
	DateTo       time.Time
}
