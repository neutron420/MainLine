package eventv1

type EventType string

const (
	EventTypeUnspecified              EventType = ""
	EventTypeSchemaVersionCreated     EventType = "schema_version_created"
	EventTypeSchemaRefreshed          EventType = "schema_refreshed"
	EventTypeMigrationStarted         EventType = "migration_started"
	EventTypeMigrationCompleted       EventType = "migration_completed"
	EventTypeMigrationFailed          EventType = "migration_failed"
	EventTypeMigrationRolledBack      EventType = "migration_rolled_back"
	EventTypeDriftDetected            EventType = "drift_detected"
	EventTypeDriftResolved            EventType = "drift_resolved"
	EventTypeConnectionCreated        EventType = "connection_created"
	EventTypeConnectionStatusChanged  EventType = "connection_status_changed"
	EventTypeMemberAdded              EventType = "member_added"
	EventTypeMemberRemoved            EventType = "member_removed"
	EventTypeRoleChanged              EventType = "role_changed"
)

type EventActor struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type EventResource struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type SchemaEvent struct {
	ID        string         `json:"id"`
	Type      string         `json:"type"`
	Version   int32          `json:"version"`
	Timestamp string         `json:"timestamp"`
	ProjectID string         `json:"project_id"`
	Actor     *EventActor    `json:"actor,omitempty"`
	Resource  *EventResource `json:"resource,omitempty"`
	Payload   string         `json:"payload,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

type SubscribeRequest struct {
	ProjectIDs  []string `json:"project_ids"`
	EventTypes  []string `json:"event_types"`
	LastEventID string   `json:"last_event_id"`
}

type SubscribeResponse struct {
	Event *SchemaEvent `json:"event"`
}

type AcknowledgeEventRequest struct {
	EventID string `json:"event_id"`
}

type AcknowledgeEventResponse struct{}

type HeartbeatRequest struct {
	ProjectIDs []string `json:"project_ids"`
}

type HeartbeatResponse struct{}
