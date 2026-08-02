package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

type TableInfo struct {
	Name    string        `json:"name"`
	Schema  string        `json:"schema"`
	Columns []ColumnInfo  `json:"columns"`
	Indexes []IndexInfo   `json:"indexes"`
	Constr  ConstraintSet `json:"constraints"`
}

type ColumnInfo struct {
	Name            string      `json:"name"`
	DataType        string      `json:"data_type"`
	IsNullable      bool        `json:"is_nullable"`
	Default         interface{} `json:"default"`
	CharMaxLength   *int        `json:"character_maximum_length,omitempty"`
	OrdinalPosition int         `json:"ordinal_position"`
}

type IndexInfo struct {
	Name      string   `json:"name"`
	Columns   []string `json:"columns"`
	Unique    bool     `json:"unique"`
	IndexType string   `json:"index_type"`
}

type ConstraintSet struct {
	PrimaryKey  []string       `json:"primary_key"`
	ForeignKeys []FKConstraint `json:"foreign_keys"`
	Uniques     []string       `json:"uniques"`
}

type FKConstraint struct {
	Column    string `json:"column"`
	RefTable  string `json:"ref_table"`
	RefColumn string `json:"ref_column"`
	Name      string `json:"name"`
}

type EnumInfo struct {
	Name   string   `json:"name"`
	Values []string `json:"values"`
}

type ExtensionInfo struct {
	Name string `json:"name"`
}

type SchemaMetadata struct {
	Tables     []TableInfo     `json:"tables"`
	Enums      []EnumInfo      `json:"enums"`
	Extensions []ExtensionInfo `json:"extensions"`
}

type IntrospectionService struct{}

func NewIntrospectionService() *IntrospectionService {
	return &IntrospectionService{}
}

func (s *IntrospectionService) Introspect(ctx context.Context, connStr string, schemaNames []string) (json.RawMessage, error) {
	pool, err := connectToDB(ctx, connStr)
	if err != nil {
		return nil, fmt.Errorf("connecting to database: %w", err)
	}
	defer pool.Close()

	meta := SchemaMetadata{
		Tables:     []TableInfo{},
		Enums:      []EnumInfo{},
		Extensions: []ExtensionInfo{},
	}

	schemaList := buildSchemaList(schemaNames)

	tables, err := s.listTables(ctx, pool, schemaList)
	if err != nil {
		return nil, fmt.Errorf("listing tables: %w", err)
	}

	for _, table := range tables {
		columns, err := s.listColumns(ctx, pool, table.Schema, table.Name)
		if err != nil {
			return nil, fmt.Errorf("listing columns for %s.%s: %w", table.Schema, table.Name, err)
		}

		indexes, err := s.listIndexes(ctx, pool, table.Schema, table.Name)
		if err != nil {
			return nil, fmt.Errorf("listing indexes for %s.%s: %w", table.Schema, table.Name, err)
		}

		constraints, err := s.listConstraints(ctx, pool, table.Schema, table.Name)
		if err != nil {
			return nil, fmt.Errorf("listing constraints for %s.%s: %w", table.Schema, table.Name, err)
		}

		meta.Tables = append(meta.Tables, TableInfo{
			Name:    table.Name,
			Schema:  table.Schema,
			Columns: columns,
			Indexes: indexes,
			Constr:  *constraints,
		})
	}

	enums, err := s.listEnums(ctx, pool, schemaList)
	if err == nil {
		meta.Enums = enums
	}

	extensions, err := s.listExtensions(ctx, pool)
	if err == nil {
		meta.Extensions = extensions
	}

	data, err := json.Marshal(meta)
	if err != nil {
		return nil, fmt.Errorf("marshaling metadata: %w", err)
	}

	return data, nil
}

func (s *IntrospectionService) listTables(ctx context.Context, pool DBPool, schemas []string) ([]struct{ Name, Schema string }, error) {
	query := `SELECT table_name, table_schema FROM information_schema.tables
		WHERE table_schema = ANY($1) AND table_type = 'BASE TABLE'
		ORDER BY table_schema, table_name`

	rows, err := pool.Query(ctx, query, schemas)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []struct{ Name, Schema string }
	for rows.Next() {
		var t struct{ Name, Schema string }
		if err := rows.Scan(&t.Name, &t.Schema); err != nil {
			return nil, err
		}
		tables = append(tables, t)
	}
	return tables, nil
}

func (s *IntrospectionService) listColumns(ctx context.Context, pool DBPool, schema, table string) ([]ColumnInfo, error) {
	query := `SELECT c.column_name, c.data_type, c.is_nullable, c.column_default, c.character_maximum_length, c.ordinal_position
		FROM information_schema.columns c
		WHERE c.table_schema = $1 AND c.table_name = $2
		ORDER BY c.ordinal_position`

	rows, err := pool.Query(ctx, query, schema, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []ColumnInfo
	for rows.Next() {
		var col ColumnInfo
		var nullable string
		var defaultVal *string
		if err := rows.Scan(&col.Name, &col.DataType, &nullable, &defaultVal, &col.CharMaxLength, &col.OrdinalPosition); err != nil {
			return nil, err
		}
		col.IsNullable = nullable == "YES"
		if defaultVal != nil {
			col.Default = *defaultVal
		}
		columns = append(columns, col)
	}
	return columns, nil
}

func (s *IntrospectionService) listIndexes(ctx context.Context, pool DBPool, schema, table string) ([]IndexInfo, error) {
	query := `SELECT i.indexname, i.indexdef FROM pg_indexes i
		WHERE i.schemaname = $1 AND i.tablename = $2`

	rows, err := pool.Query(ctx, query, schema, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var indexes []IndexInfo
	for rows.Next() {
		var name, def string
		if err := rows.Scan(&name, &def); err != nil {
			return nil, err
		}

		idx := IndexInfo{
			Name:      name,
			Unique:    strings.Contains(def, "UNIQUE"),
			IndexType: "btree",
		}

		if strings.Contains(def, "USING gin") {
			idx.IndexType = "gin"
		} else if strings.Contains(def, "USING gist") {
			idx.IndexType = "gist"
		} else if strings.Contains(def, "USING hash") {
			idx.IndexType = "hash"
		}

		if start := strings.Index(def, "("); start != -1 {
			end := strings.LastIndex(def, ")")
			if end != -1 {
				cols := strings.TrimSpace(def[start+1 : end])
				for _, c := range strings.Split(cols, ",") {
					idx.Columns = append(idx.Columns, strings.TrimSpace(c))
				}
			}
		}

		indexes = append(indexes, idx)
	}
	return indexes, nil
}

func (s *IntrospectionService) listConstraints(ctx context.Context, pool DBPool, schema, table string) (*ConstraintSet, error) {
	cs := &ConstraintSet{
		PrimaryKey:  []string{},
		ForeignKeys: []FKConstraint{},
		Uniques:     []string{},
	}

	pkQuery := `SELECT kcu.column_name FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kcu ON tc.constraint_name = kcu.constraint_name
		WHERE tc.table_schema = $1 AND tc.table_name = $2 AND tc.constraint_type = 'PRIMARY KEY'
		ORDER BY kcu.ordinal_position`

	pkRows, err := pool.Query(ctx, pkQuery, schema, table)
	if err == nil {
		defer pkRows.Close()
		for pkRows.Next() {
			var col string
			if err := pkRows.Scan(&col); err == nil {
				cs.PrimaryKey = append(cs.PrimaryKey, col)
			}
		}
	}

	fkQuery := `SELECT kcu.column_name, ccu.table_name AS ref_table, ccu.column_name AS ref_column, tc.constraint_name
		FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kcu ON tc.constraint_name = kcu.constraint_name
		JOIN information_schema.constraint_column_usage ccu ON tc.constraint_name = ccu.constraint_name
		WHERE tc.table_schema = $1 AND tc.table_name = $2 AND tc.constraint_type = 'FOREIGN KEY'`

	fkRows, err := pool.Query(ctx, fkQuery, schema, table)
	if err == nil {
		defer fkRows.Close()
		for fkRows.Next() {
			var fk FKConstraint
			if err := fkRows.Scan(&fk.Column, &fk.RefTable, &fk.RefColumn, &fk.Name); err == nil {
				cs.ForeignKeys = append(cs.ForeignKeys, fk)
			}
		}
	}

	uqQuery := `SELECT kcu.column_name FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kcu ON tc.constraint_name = kcu.constraint_name
		WHERE tc.table_schema = $1 AND tc.table_name = $2 AND tc.constraint_type = 'UNIQUE'
		ORDER BY kcu.ordinal_position`

	uqRows, err := pool.Query(ctx, uqQuery, schema, table)
	if err == nil {
		defer uqRows.Close()
		for uqRows.Next() {
			var col string
			if err := uqRows.Scan(&col); err == nil {
				cs.Uniques = append(cs.Uniques, col)
			}
		}
	}

	return cs, nil
}

func (s *IntrospectionService) listEnums(ctx context.Context, pool DBPool, schemas []string) ([]EnumInfo, error) {
	query := `SELECT t.typname, array_agg(e.enumlabel ORDER BY e.enumsortorder) as values
		FROM pg_type t
		JOIN pg_enum e ON t.oid = e.enumtypid
		JOIN pg_namespace n ON t.typnamespace = n.oid
		WHERE n.nspname = ANY($1)
		GROUP BY t.typname`

	rows, err := pool.Query(ctx, query, schemas)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var enums []EnumInfo
	for rows.Next() {
		var e EnumInfo
		if err := rows.Scan(&e.Name, &e.Values); err != nil {
			return nil, err
		}
		enums = append(enums, e)
	}
	return enums, nil
}

func (s *IntrospectionService) listExtensions(ctx context.Context, pool DBPool) ([]ExtensionInfo, error) {
	rows, err := pool.Query(ctx, "SELECT extname FROM pg_extension ORDER BY extname")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var extensions []ExtensionInfo
	for rows.Next() {
		var e ExtensionInfo
		if err := rows.Scan(&e.Name); err != nil {
			return nil, err
		}
		extensions = append(extensions, e)
	}
	return extensions, nil
}

func buildSchemaList(names []string) []string {
	if len(names) == 0 {
		return []string{"public"}
	}
	return names
}

// DBPool interface for introspection queries
type DBPool interface {
	Query(ctx context.Context, sql string, args ...any) (Rows, error)
	Close()
}

type Rows interface {
	Next() bool
	Scan(dest ...interface{}) error
	Close()
}

// connectToDB is overridden in tests; in production uses pgxpool
var connectToDB = func(ctx context.Context, connStr string) (DBPool, error) {
	return connectPGX(ctx, connStr)
}

func connectPGX(ctx context.Context, connStr string) (DBPool, error) {
	pool, err := openPool(ctx, connStr)
	if err != nil {
		return nil, err
	}
	return pool, nil
}

var openPool = func(ctx context.Context, connStr string) (DBPool, error) {
	return nil, fmt.Errorf("not implemented: use setConnector")
}

func SetConnector(fn func(ctx context.Context, connStr string) (DBPool, error)) {
	openPool = fn
	connectToDB = func(ctx context.Context, connStr string) (DBPool, error) {
		return fn(ctx, connStr)
	}
}
