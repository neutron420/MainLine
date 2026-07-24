package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/schemahub/backend/pkg/encryption"
)

type ConnectionService struct {
	connRepo      ConnectionRepository
	encryptionKey []byte
}

func NewConnectionService(connRepo ConnectionRepository, encryptionKey []byte) *ConnectionService {
	return &ConnectionService{connRepo: connRepo, encryptionKey: encryptionKey}
}

func (s *ConnectionService) Create(ctx context.Context, c *Connection, rawPassword string) (*Connection, error) {
	encrypted, err := encryption.Encrypt([]byte(rawPassword), s.encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("encrypting password: %w", err)
	}
	c.PasswordEncrypted = encrypted
	c.ConnectionStatus = ConnStatusUnknown

	if err := s.connRepo.Create(ctx, c); err != nil {
		return nil, fmt.Errorf("creating connection: %w", err)
	}

	created, err := s.connRepo.GetByID(ctx, c.ID)
	if err != nil {
		return nil, fmt.Errorf("retrieving connection: %w", err)
	}

	return created, nil
}

func (s *ConnectionService) GetByID(ctx context.Context, id string) (*Connection, error) {
	return s.connRepo.GetByID(ctx, id)
}

func (s *ConnectionService) List(ctx context.Context, projectID, cursor string, pageSize int32) ([]*Connection, string, int32, error) {
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	return s.connRepo.ListByProjectID(ctx, projectID, cursor, pageSize)
}

func (s *ConnectionService) Update(ctx context.Context, id, name, host string, port int32, databaseName, username, password, sslMode string) (*Connection, error) {
	c, err := s.connRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("connection not found: %w", err)
	}

	if name != "" {
		c.Name = name
	}
	if host != "" {
		c.Host = host
	}
	if port > 0 {
		c.Port = int(port)
	}
	if databaseName != "" {
		c.DatabaseName = databaseName
	}
	if username != "" {
		c.Username = username
	}
	if password != "" {
		encrypted, err := encryption.Encrypt([]byte(password), s.encryptionKey)
		if err != nil {
			return nil, fmt.Errorf("encrypting password: %w", err)
		}
		c.PasswordEncrypted = encrypted
	}
	if sslMode != "" {
		c.SSLMode = SSLMode(sslMode)
	}

	if err := s.connRepo.Update(ctx, c); err != nil {
		return nil, fmt.Errorf("updating connection: %w", err)
	}

	return s.connRepo.GetByID(ctx, id)
}

func (s *ConnectionService) GetConnectionString(ctx context.Context, id string) (string, error) {
	c, err := s.connRepo.GetByID(ctx, id)
	if err != nil {
		return "", fmt.Errorf("connection not found: %w", err)
	}

	decrypted, err := encryption.Decrypt(c.PasswordEncrypted, s.encryptionKey)
	if err != nil {
		return "", fmt.Errorf("decrypting password: %w", err)
	}

	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.Username, string(decrypted), c.Host, c.Port, c.DatabaseName, string(c.SSLMode)), nil
}

func (s *ConnectionService) Delete(ctx context.Context, id string) error {
	return s.connRepo.SoftDelete(ctx, id)
}

func (s *ConnectionService) Test(ctx context.Context, connectionID string) (bool, int32, string, string, error) {
	c, err := s.connRepo.GetByID(ctx, connectionID)
	if err != nil {
		return false, 0, "", "", fmt.Errorf("connection not found")
	}

	decrypted, err := encryption.Decrypt(c.PasswordEncrypted, s.encryptionKey)
	if err != nil {
		return false, 0, "", "", fmt.Errorf("decrypting password: %w", err)
	}

	start := time.Now()
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s", c.Username, string(decrypted), c.Host, c.Port, c.DatabaseName, string(c.SSLMode))

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		s.connRepo.UpdateStatus(ctx, connectionID, ConnStatusFailed, nil)
		latency := int32(time.Since(start).Milliseconds())
		return false, latency, "", "", fmt.Errorf("connection failed: %w", err)
	}
	defer pool.Close()

	var serverVersion string
	var dbName string
	if err := pool.QueryRow(ctx, "SELECT version(), current_database()").Scan(&serverVersion, &dbName); err != nil {
		s.connRepo.UpdateStatus(ctx, connectionID, ConnStatusFailed, nil)
		latency := int32(time.Since(start).Milliseconds())
		return false, latency, serverVersion, "", fmt.Errorf("query failed: %w", err)
	}

	latency := int32(time.Since(start).Milliseconds())
	now := time.Now()
	s.connRepo.UpdateStatus(ctx, connectionID, ConnStatusConnected, &now)

	return true, latency, serverVersion, dbName, nil
}
