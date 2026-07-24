package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type SchemaCache struct {
	rdb *redis.Client
}

func NewSchemaCache(rdb *redis.Client) *SchemaCache {
	return &SchemaCache{rdb: rdb}
}

func (c *SchemaCache) GetVersion(ctx context.Context, versionID string) (*SchemaVersion, error) {
	data, err := c.rdb.Get(ctx, cacheKey("version", versionID)).Bytes()
	if err != nil {
		return nil, nil
	}
	var v SchemaVersion
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, nil
	}
	return &v, nil
}

func (c *SchemaCache) SetVersion(ctx context.Context, v *SchemaVersion) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, cacheKey("version", v.ID), data, 5*time.Minute).Err()
}

func (c *SchemaCache) InvalidateVersion(ctx context.Context, versionID string) error {
	return c.rdb.Del(ctx, cacheKey("version", versionID)).Err()
}

func (c *SchemaCache) GetSchema(ctx context.Context, schemaID string) (*Schema, error) {
	data, err := c.rdb.Get(ctx, cacheKey("schema", schemaID)).Bytes()
	if err != nil {
		return nil, nil
	}
	var s Schema
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, nil
	}
	return &s, nil
}

func (c *SchemaCache) SetSchema(ctx context.Context, s *Schema) error {
	data, err := json.Marshal(s)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, cacheKey("schema", s.ID), data, 5*time.Minute).Err()
}

func (c *SchemaCache) InvalidateSchema(ctx context.Context, schemaID string) error {
	return c.rdb.Del(ctx, cacheKey("schema", schemaID)).Err()
}

func (c *SchemaCache) GetDiagram(ctx context.Context, versionID string) (*DiagramData, error) {
	data, err := c.rdb.Get(ctx, cacheKey("diagram", versionID)).Bytes()
	if err != nil {
		return nil, nil
	}
	var d DiagramData
	if err := json.Unmarshal(data, &d); err != nil {
		return nil, nil
	}
	return &d, nil
}

func (c *SchemaCache) SetDiagram(ctx context.Context, versionID string, diagram *DiagramData) error {
	data, err := json.Marshal(diagram)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, cacheKey("diagram", versionID), data, 10*time.Minute).Err()
}

func cacheKey(parts ...string) string {
	return fmt.Sprintf("schema:%s", join(parts, ":"))
}

func join(parts []string, sep string) string {
	result := parts[0]
	for _, p := range parts[1:] {
		result += sep + p
	}
	return result
}
