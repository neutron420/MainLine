package domain

import "time"

type EventType string

const (
	EventTypeSchemaVersionCreated    EventType = "schema_version_created"
	EventTypeSchemaRefreshed         EventType = "schema_refreshed"
	EventTypeMigrationStarted        EventType = "migration_started"
	EventTypeMigrationCompleted      EventType = "migration_completed"
	EventTypeMigrationFailed         EventType = "migration_failed"
	EventTypeMigrationRolledBack     EventType = "migration_rolled_back"
	EventTypeDriftDetected           EventType = "drift_detected"
	EventTypeDriftResolved           EventType = "drift_resolved"
	EventTypeConnectionCreated       EventType = "connection_created"
	EventTypeConnectionStatusChanged EventType = "connection_status_changed"
	EventTypeMemberAdded             EventType = "member_added"
	EventTypeMemberRemoved           EventType = "member_removed"
	EventTypeRoleChanged             EventType = "role_changed"
)

type EventActor struct {
	ID    string
	Email string
}

type EventResource struct {
	Type string
	ID   string
}

type SchemaEvent struct {
	ID        string
	Type      EventType
	Version   int32
	Timestamp time.Time
	ProjectID string
	Actor     *EventActor
	Resource  *EventResource
	Payload   string
	Metadata  map[string]string
}

type Subscriber struct {
	ID         string
	UserID     string
	ProjectIDs []string
	EventTypes []EventType
	Buffer     chan *SchemaEvent
	Done       chan struct{}
}

type PresenceEntry struct {
	UserID    string
	ProjectID string
	Timestamp time.Time
}
