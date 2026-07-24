package domain

import (
	"context"
	"fmt"
	"time"
)

type ConnectionStatus string

const (
	ConnStatusUnknown  ConnectionStatus = "unknown"
	ConnStatusConnected ConnectionStatus = "connected"
	ConnStatusFailed   ConnectionStatus = "failed"
)

type SSLMode string

const (
	SSLDisable    SSLMode = "disable"
	SSLAllow      SSLMode = "allow"
	SSLPrefer     SSLMode = "prefer"
	SSLRequire    SSLMode = "require"
	SSLVerifyCA   SSLMode = "verify-ca"
	SSLVerifyFull SSLMode = "verify-full"
)

type Connection struct {
	ID                string
	ProjectID         string
	Name              string
	Host              string
	Port              int
	DatabaseName      string
	Username          string
	PasswordEncrypted string
	SSLMode           SSLMode
	ConnectionStatus  ConnectionStatus
	LastConnectedAt   *time.Time
	CreatedBy         string
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         *time.Time
}

type ConnectionRepository interface {
	Create(ctx context.Context, c *Connection) error
	GetByID(ctx context.Context, id string) (*Connection, error)
	ListByProjectID(ctx context.Context, projectID, cursor string, limit int32) ([]*Connection, string, int32, error)
	Update(ctx context.Context, c *Connection) error
	SoftDelete(ctx context.Context, id string) error
	UpdateStatus(ctx context.Context, id string, status ConnectionStatus, lastConnectedAt *time.Time) error
}

func (c *Connection) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("connection name is required")
	}
	if c.Host == "" {
		return fmt.Errorf("host is required")
	}
	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}
	if c.DatabaseName == "" {
		return fmt.Errorf("database name is required")
	}
	if c.Username == "" {
		return fmt.Errorf("username is required")
	}
	if c.SSLMode == "" {
		c.SSLMode = SSLRequire
	}
	return nil
}
