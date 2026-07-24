package auditv1

type AuditEntry struct {
	ID            string            `json:"id"`
	EventType     string            `json:"event_type"`
	ActorID       string            `json:"actor_id,omitempty"`
	ActorEmail    string            `json:"actor_email,omitempty"`
	Action        string            `json:"action"`
	ResourceType  string            `json:"resource_type"`
	ResourceID    string            `json:"resource_id"`
	ResourceChanges string          `json:"resource_changes,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	IPAddress     string            `json:"ip_address,omitempty"`
	UserAgent     string            `json:"user_agent,omitempty"`
	TraceID       string            `json:"trace_id"`
	CreatedAt     string            `json:"created_at"`
}

type ListAuditEntriesRequest struct {
	EventType string `json:"event_type,omitempty"`
	ActorID   string `json:"actor_id,omitempty"`
	ResourceType string `json:"resource_type,omitempty"`
	ResourceID   string `json:"resource_id,omitempty"`
	DateFrom string `json:"date_from,omitempty"`
	DateTo   string `json:"date_to,omitempty"`
	Cursor   string `json:"cursor"`
	PageSize int32  `json:"page_size"`
}

type ListAuditEntriesResponse struct {
	Entries    []*AuditEntry `json:"entries"`
	NextCursor string        `json:"next_cursor"`
	TotalCount int32         `json:"total_count"`
}

type GetAuditEntryRequest struct {
	ID string `json:"id"`
}

type GetAuditEntryResponse struct {
	Entry *AuditEntry `json:"entry"`
}

type TailAuditEntriesRequest struct {
	EventType    string `json:"event_type,omitempty"`
	SinceEventID string `json:"since_event_id,omitempty"`
}

type GetAuditStatsRequest struct {
	DateFrom string `json:"date_from,omitempty"`
	DateTo   string `json:"date_to,omitempty"`
}

type GetAuditStatsResponse struct {
	TotalEntries  int32              `json:"total_entries"`
	ByEventType   map[string]int32   `json:"by_event_type"`
	ByAction      map[string]int32   `json:"by_action"`
	UniqueActors  int32              `json:"unique_actors"`
	DateFrom      string             `json:"date_from"`
	DateTo        string             `json:"date_to"`
}
