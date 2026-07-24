package driftv1

type DriftEvent struct {
	ID                 string `json:"id"`
	ConnectionID       string `json:"connection_id"`
	SchemaID           string `json:"schema_id,omitempty"`
	ExpectedVersionID  string `json:"expected_version_id,omitempty"`
	DriftType          string `json:"drift_type"`
	ObjectType         string `json:"object_type"`
	ObjectName         string `json:"object_name"`
	ExpectedDefinition string `json:"expected_definition,omitempty"`
	ActualDefinition   string `json:"actual_definition,omitempty"`
	DiffSummary        string `json:"diff_summary,omitempty"`
	Severity           string `json:"severity"`
	Status             string `json:"status"`
	DetectedAt         string `json:"detected_at"`
	ResolvedAt         string `json:"resolved_at,omitempty"`
	ResolvedBy         string `json:"resolved_by,omitempty"`
}

type CheckDriftRequest struct {
	ConnectionID string   `json:"connection_id"`
	SchemaNames  []string `json:"schema_names,omitempty"`
}

type CheckDriftResponse struct {
	Events      []*DriftEvent `json:"events,omitempty"`
	HasDrift    bool          `json:"has_drift"`
	TotalDrifts int32         `json:"total_drifts"`
}

type ListDriftEventsRequest struct {
	ConnectionID string `json:"connection_id,omitempty"`
	Status       string `json:"status,omitempty"`
	Severity     string `json:"severity,omitempty"`
	DriftType    string `json:"drift_type,omitempty"`
	Cursor       string `json:"cursor"`
	PageSize     int32  `json:"page_size"`
}

type ListDriftEventsResponse struct {
	Events     []*DriftEvent `json:"events"`
	NextCursor string        `json:"next_cursor"`
	TotalCount int32         `json:"total_count"`
}

type GetDriftEventRequest struct {
	ID string `json:"id"`
}

type GetDriftEventResponse struct {
	Event *DriftEvent `json:"event"`
}

type ResolveDriftEventRequest struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type ResolveDriftEventResponse struct {
	Event *DriftEvent `json:"event"`
}

type GetDriftStatsRequest struct {
	ConnectionID string `json:"connection_id,omitempty"`
}

type GetDriftStatsResponse struct {
	TotalOpen       int32            `json:"total_open"`
	TotalResolved   int32            `json:"total_resolved"`
	TotalAcknowledged int32          `json:"total_acknowledged"`
	TotalFalsePositive int32         `json:"total_false_positive"`
	BySeverity      map[string]int32 `json:"by_severity"`
	ByDriftType     map[string]int32 `json:"by_drift_type"`
}
