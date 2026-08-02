package domain

import (
	"strings"
	"testing"
)

func TestSQLValidator_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		upSQL   string
		downSQL string
		wantErr string
	}{
		{
			name:    "empty up sql",
			upSQL:   "",
			wantErr: "up_sql is required",
		},
		{
			name:    "valid create table",
			upSQL:   "CREATE TABLE users (id uuid PRIMARY KEY);",
			wantErr: "",
		},
		{
			name:    "valid alter with semicolon",
			upSQL:   "ALTER TABLE users ADD COLUMN email varchar(255);",
			wantErr: "",
		},
		{
			name:    "missing semicolon",
			upSQL:   "CREATE TABLE users (id uuid PRIMARY KEY)",
			wantErr: "must end with semicolon",
		},
		{
			name:    "drop database banned",
			upSQL:   "DROP DATABASE mainline;",
			wantErr: "DROP DATABASE",
		},
		{
			name:    "alter system banned",
			upSQL:   "ALTER SYSTEM SET shared_buffers = '1GB';",
			wantErr: "ALTER SYSTEM",
		},
		{
			name:    "drop extension banned",
			upSQL:   "DROP EXTENSION pgcrypto;",
			wantErr: "DROP EXTENSION",
		},
		{
			name:    "drop table without if exists",
			upSQL:   "DROP TABLE users;",
			wantErr: "DROP TABLE without IF EXISTS",
		},
		{
			name:    "drop table with if exists is allowed",
			upSQL:   "DROP TABLE IF EXISTS users;",
			wantErr: "",
		},
		{
			name:    "unrecognized verb",
			upSQL:   "FROBNICATE users;",
			wantErr: "unrecognized SQL command",
		},
		{
			name:    "multiple statements with comment",
			upSQL:   "-- add email\nALTER TABLE users ADD COLUMN email varchar(255);\nCREATE INDEX idx_users_email ON users(email);",
			wantErr: "",
		},
		{
			name:    "semicolon inside string literal",
			upSQL:   `INSERT INTO meta (k) VALUES ('a;b');`,
			wantErr: "",
		},
		{
			name:    "down sql validated too",
			upSQL:   "CREATE TABLE users (id uuid);",
			downSQL: "DROP DATABASE mainline;",
			wantErr: "DROP DATABASE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewSQLValidator()
			ok, errs := v.Validate(tt.upSQL, tt.downSQL)

			if tt.wantErr == "" {
				if !ok {
					t.Errorf("Validate() = %v, want valid; errors: %v", ok, errs)
				}
				return
			}

			if ok {
				t.Fatalf("Validate() = valid, want error containing %q", tt.wantErr)
			}
			found := false
			for _, e := range errs {
				if strings.Contains(e, tt.wantErr) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("errors %v do not contain %q", errs, tt.wantErr)
			}
		})
	}
}

func TestSQLValidator_ValidateWithWarnings(t *testing.T) {
	t.Parallel()

	t.Run("alter drop column valid but warns", func(t *testing.T) {
		v := NewSQLValidator()
		res := v.ValidateWithWarnings("ALTER TABLE users DROP COLUMN zip;", "")
		if !res.Valid {
			t.Errorf("Valid = false, want true; errors: %v", res.Errors)
		}
		if len(res.Warnings) != 1 {
			t.Errorf("warnings len = %d, want 1", len(res.Warnings))
		}
	})

	t.Run("drop index warns but valid", func(t *testing.T) {
		v := NewSQLValidator()
		res := v.ValidateWithWarnings("DROP INDEX idx_users_email;", "")
		if !res.Valid {
			t.Errorf("Valid = false, want true; errors: %v", res.Errors)
		}
		if len(res.Warnings) == 0 {
			t.Error("expected warning for DROP INDEX")
		}
	})

	t.Run("empty up sql produces error", func(t *testing.T) {
		v := NewSQLValidator()
		res := v.ValidateWithWarnings("", "")
		if res.Valid {
			t.Error("Valid = true, want false")
		}
		if len(res.Errors) != 1 {
			t.Errorf("errors len = %d, want 1", len(res.Errors))
		}
	})
}

func TestSplitStatements(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		sql  string
		want int
	}{
		{name: "single", sql: "CREATE TABLE a (id int);", want: 1},
		{name: "two", sql: "CREATE TABLE a (id int); CREATE TABLE b (id int);", want: 2},
		{name: "comment stripped", sql: "CREATE TABLE a (id int); -- comment\nCREATE TABLE b (id int);", want: 2},
		{name: "semicolon in parens", sql: "CREATE FUNCTION f() RETURNS int AS 'SELECT 1;' LANGUAGE sql;", want: 1},
		{name: "empty", sql: "  ", want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitStatements(tt.sql)
			if len(got) != tt.want {
				t.Errorf("splitStatements() = %d statements, want %d (got %q)", len(got), tt.want, got)
			}
		})
	}
}

func TestParseStatement(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		sql        string
		wantVerb   string
		wantObject string
		wantName   string
	}{
		{name: "create table", sql: "CREATE TABLE users (id int)", wantVerb: "CREATE", wantObject: "TABLE", wantName: "USERS"},
		{name: "alter table", sql: "ALTER TABLE users ADD COLUMN email text", wantVerb: "ALTER", wantObject: "TABLE", wantName: "USERS"},
		{name: "drop index", sql: "DROP INDEX idx_users_email", wantVerb: "DROP", wantObject: "INDEX", wantName: "IDX_USERS_EMAIL"},
		{name: "create index", sql: "CREATE INDEX idx ON users (email)", wantVerb: "CREATE", wantObject: "INDEX", wantName: "IDX"},
		{name: "if not exists create", sql: "CREATE TABLE IF NOT EXISTS sessions (id uuid)", wantVerb: "CREATE", wantObject: "TABLE", wantName: "IF"},
		{name: "select", sql: "SELECT * FROM users", wantVerb: "SELECT"},
		{name: "empty", sql: "", wantVerb: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := parseStatement(tt.sql)
			if info.verb != tt.wantVerb {
				t.Errorf("verb = %q, want %q", info.verb, tt.wantVerb)
			}
			if info.object != tt.wantObject {
				t.Errorf("object = %q, want %q", info.object, tt.wantObject)
			}
			if info.name != tt.wantName {
				t.Errorf("name = %q, want %q", info.name, tt.wantName)
			}
		})
	}
}
