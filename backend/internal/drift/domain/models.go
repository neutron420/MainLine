package domain

import "time"

type DriftType string

const (
	DriftTypeMissingObject  DriftType = "missing_object"
	DriftTypeExtraObject    DriftType = "extra_object"
	DriftTypeModifiedObject DriftType = "modified_object"
	DriftTypeTypeChange     DriftType = "type_change"
)

type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityCritical Severity = "critical"
)

type DriftStatus string

const (
	DriftStatusOpen          DriftStatus = "open"
	DriftStatusAcknowledged  DriftStatus = "acknowledged"
	DriftStatusResolved      DriftStatus = "resolved"
	DriftStatusFalsePositive DriftStatus = "false_positive"
)

type DriftEvent struct {
	ID                 string
	ConnectionID       string
	SchemaID           string
	ExpectedVersionID  string
	DriftType          DriftType
	ObjectType         string
	ObjectName         string
	ExpectedDefinition string
	ActualDefinition   string
	DiffSummary        string
	Severity           Severity
	Status             DriftStatus
	DetectedAt         time.Time
	ResolvedAt         *time.Time
	ResolvedBy         string
}

type DriftStats struct {
	TotalOpen          int32
	TotalResolved      int32
	TotalAcknowledged  int32
	TotalFalsePositive int32
	BySeverity         map[string]int32
	ByDriftType        map[string]int32
}
