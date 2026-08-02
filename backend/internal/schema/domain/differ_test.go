package domain

import (
	"encoding/json"
	"reflect"
	"testing"
)

func mustJSON(t *testing.T, v any) json.RawMessage {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return b
}

func TestComputeChecksum(t *testing.T) {
	t.Parallel()

	a := json.RawMessage(`{"tables":[]}`)
	b := json.RawMessage(`{"tables":[]}`)
	c := json.RawMessage(`{"tables":[{"name":"x"}]}`)

	ca := ComputeChecksum(a)
	if ca == "" {
		t.Fatal("ComputeChecksum returned empty")
	}
	if ca != ComputeChecksum(b) {
		t.Error("identical content produced different checksums")
	}
	if ca == ComputeChecksum(c) {
		t.Error("different content produced same checksum")
	}
	if len(ca) != 64 {
		t.Errorf("checksum length = %d, want 64 (sha256 hex)", len(ca))
	}
}

func buildTable(schema, name string, columns []ColumnInfo, indexes []IndexInfo, fks []FKConstraint) TableInfo {
	return TableInfo{
		Schema:  schema,
		Name:    name,
		Columns: columns,
		Indexes: indexes,
		Constr: ConstraintSet{
			ForeignKeys: fks,
		},
	}
}

func TestDiffer_IdenticalSchemas(t *testing.T) {
	t.Parallel()

	table := buildTable("public", "users", []ColumnInfo{
		{Name: "id", DataType: "uuid", IsNullable: false},
		{Name: "email", DataType: "varchar(255)", IsNullable: false},
	}, []IndexInfo{
		{Name: "idx_users_email", Columns: []string{"email"}},
	}, nil)

	metaA := SchemaMetadata{Tables: []TableInfo{table}, Enums: []EnumInfo{{Name: "status", Values: []string{"active", "inactive"}}}}
	metaB := SchemaMetadata{Tables: []TableInfo{table}, Enums: []EnumInfo{{Name: "status", Values: []string{"active", "inactive"}}}}

	d := NewDiffer()
	diff, err := d.Diff(mustJSON(t, metaA), mustJSON(t, metaB))
	if err != nil {
		t.Fatalf("Diff() error = %v", err)
	}

	if len(diff.AddedObjects) != 0 || len(diff.RemovedObjects) != 0 || len(diff.ModifiedObjects) != 0 {
		t.Errorf("identical schemas produced diff: %+v", diff)
	}
}

func TestDiffer_AddedAndRemovedTables(t *testing.T) {
	t.Parallel()

	users := buildTable("public", "users", []ColumnInfo{{Name: "id", DataType: "uuid"}}, nil, nil)
	teams := buildTable("public", "teams", []ColumnInfo{{Name: "id", DataType: "uuid"}}, nil, nil)

	metaA := SchemaMetadata{Tables: []TableInfo{users}}
	metaB := SchemaMetadata{Tables: []TableInfo{users, teams}}

	d := NewDiffer()
	diff, err := d.Diff(mustJSON(t, metaA), mustJSON(t, metaB))
	if err != nil {
		t.Fatalf("Diff() error = %v", err)
	}

	if len(diff.AddedObjects) != 1 {
		t.Fatalf("AddedObjects len = %d, want 1", len(diff.AddedObjects))
	}
	if diff.AddedObjects[0].Name != "public.teams" {
		t.Errorf("added object = %q, want public.teams", diff.AddedObjects[0].Name)
	}
	if diff.AddedObjects[0].Type != "table" {
		t.Errorf("added object type = %q, want table", diff.AddedObjects[0].Type)
	}
	if len(diff.RemovedObjects) != 0 {
		t.Errorf("RemovedObjects len = %d, want 0", len(diff.RemovedObjects))
	}
}

func TestDiffer_ModifiedColumnTypeAndNullable(t *testing.T) {
	t.Parallel()

	colA := ColumnInfo{Name: "status", DataType: "varchar(20)", IsNullable: true}
	colB := ColumnInfo{Name: "status", DataType: "enum", IsNullable: false}

	metaA := SchemaMetadata{Tables: []TableInfo{buildTable("public", "users", []ColumnInfo{colA}, nil, nil)}}
	metaB := SchemaMetadata{Tables: []TableInfo{buildTable("public", "users", []ColumnInfo{colB}, nil, nil)}}

	d := NewDiffer()
	diff, err := d.Diff(mustJSON(t, metaA), mustJSON(t, metaB))
	if err != nil {
		t.Fatal(err)
	}

	if len(diff.ModifiedObjects) != 1 {
		t.Fatalf("ModifiedObjects len = %d, want 1", len(diff.ModifiedObjects))
	}

	changes := diff.ModifiedObjects[0].Changes
	if len(changes) != 2 {
		t.Fatalf("changes len = %d, want 2 (type + nullable)", len(changes))
	}

	found := map[string]bool{}
	for _, c := range changes {
		found[c.Field] = true
	}
	if !found["users.status.type"] {
		t.Errorf("missing type change, got %+v", changes)
	}
	if !found["users.status.nullable"] {
		t.Errorf("missing nullable change, got %+v", changes)
	}
}

func TestDiffer_IndexChanges(t *testing.T) {
	t.Parallel()

	idxA := []IndexInfo{{Name: "idx_users_email", Columns: []string{"email"}}}
	idxB := []IndexInfo{{Name: "idx_users_email", Columns: []string{"email"}}, {Name: "idx_users_name", Columns: []string{"name"}}}

	metaA := SchemaMetadata{Tables: []TableInfo{buildTable("public", "users", nil, idxA, nil)}}
	metaB := SchemaMetadata{Tables: []TableInfo{buildTable("public", "users", nil, idxB, nil)}}

	d := NewDiffer()
	diff, err := d.Diff(mustJSON(t, metaA), mustJSON(t, metaB))
	if err != nil {
		t.Fatal(err)
	}

	if len(diff.ModifiedObjects) != 1 {
		t.Fatalf("ModifiedObjects len = %d, want 1", len(diff.ModifiedObjects))
	}
	changes := diff.ModifiedObjects[0].Changes
	if len(changes) != 1 || changes[0].Field != "index.idx_users_name.added" {
		t.Errorf("expected index added change, got %+v", changes)
	}
}

func TestDiffer_EnumChanges(t *testing.T) {
	t.Parallel()

	enumA := EnumInfo{Name: "status", Values: []string{"active", "inactive"}}
	enumB := EnumInfo{Name: "status", Values: []string{"active", "inactive", "suspended"}}
	enumC := EnumInfo{Name: "priority", Values: []string{"low"}}

	metaA := SchemaMetadata{Enums: []EnumInfo{enumA}}
	metaB := SchemaMetadata{Enums: []EnumInfo{enumB, enumC}}

	d := NewDiffer()
	diff, err := d.Diff(mustJSON(t, metaA), mustJSON(t, metaB))
	if err != nil {
		t.Fatal(err)
	}

	if len(diff.AddedObjects) != 1 || diff.AddedObjects[0].Name != "priority" {
		t.Errorf("expected enum priority added, got %+v", diff.AddedObjects)
	}
}

func TestDiffer_InvalidJSON(t *testing.T) {
	t.Parallel()

	d := NewDiffer()
	if _, err := d.Diff(json.RawMessage("not json"), mustJSON(t, SchemaMetadata{})); err == nil {
		t.Error("Diff with invalid JSON = nil error, want error")
	}
	if _, err := d.Diff(mustJSON(t, SchemaMetadata{}), json.RawMessage("not json")); err == nil {
		t.Error("Diff with invalid JSON (B) = nil error, want error")
	}
}

func TestDetectBreakingChanges(t *testing.T) {
	t.Parallel()

	d := NewDiffer()

	t.Run("removed table", func(t *testing.T) {
		diff := &DiffResult{
			RemovedObjects: []DiffObject{{Type: "table", Name: "public.legacy"}},
		}
		changes := d.DetectBreakingChanges(diff)
		if len(changes) != 1 || changes[0].Severity != "breaking" || changes[0].Change != "removed" {
			t.Errorf("unexpected changes: %+v", changes)
		}
	})

	t.Run("removed column and type change", func(t *testing.T) {
		diff := &DiffResult{
			ModifiedObjects: []DiffObject{{
				Type: "table", Name: "public.users",
				Changes: []FieldChange{
					{Field: "users.zip.removed", Before: "varchar(10)", After: nil},
					{Field: "users.status.type", Before: "varchar(20)", After: "enum"},
				},
			}},
		}
		changes := d.DetectBreakingChanges(diff)
		if len(changes) != 2 {
			t.Fatalf("changes len = %d, want 2", len(changes))
		}
		for _, c := range changes {
			if c.Severity != "breaking" {
				t.Errorf("change %s severity = %s, want breaking", c.ObjectName, c.Severity)
			}
		}
	})

	t.Run("nullable false to true is caution", func(t *testing.T) {
		diff := &DiffResult{
			ModifiedObjects: []DiffObject{{
				Type: "table", Name: "public.users",
				Changes: []FieldChange{
					{Field: "users.phone.nullable", Before: false, After: true},
				},
			}},
		}
		changes := d.DetectBreakingChanges(diff)
		if len(changes) != 1 || changes[0].Severity != "caution" {
			t.Errorf("expected caution change, got %+v", changes)
		}
	})

	t.Run("nullable true to false is breaking", func(t *testing.T) {
		diff := &DiffResult{
			ModifiedObjects: []DiffObject{{
				Type: "table", Name: "public.users",
				Changes: []FieldChange{
					{Field: "users.phone.nullable", Before: true, After: false},
				},
			}},
		}
		changes := d.DetectBreakingChanges(diff)
		if len(changes) != 1 || changes[0].Severity != "breaking" {
			t.Errorf("expected breaking change, got %+v", changes)
		}
	})

	t.Run("malformed field names skipped", func(t *testing.T) {
		diff := &DiffResult{
			ModifiedObjects: []DiffObject{{
				Type: "table", Name: "public.users",
				Changes: []FieldChange{{Field: "no-dots", Before: nil, After: "x"}},
			}},
		}
		if changes := d.DetectBreakingChanges(diff); len(changes) != 0 {
			t.Errorf("expected no changes, got %+v", changes)
		}
	})
}

func TestAnalyzeImpact(t *testing.T) {
	t.Parallel()

	d := NewDiffer()

	meta := impactMetadata{
		Tables: []TableInfo{
			buildTable("public", "users", nil, nil, []FKConstraint{{Name: "fk_users_team", RefTable: "teams"}}),
			buildTable("public", "teams", nil, nil, nil),
		},
		Views: []viewInfo{
			{Name: "active_teams", Schema: "public", Definition: "SELECT * FROM public.teams WHERE active"},
		},
	}

	breaking := []BreakingChange{{ObjectType: "table", ObjectName: "public.teams", Change: "removed"}}
	impacted := d.AnalyzeImpact(mustJSON(t, meta), breaking)

	if len(impacted) != 2 {
		t.Fatalf("impacted len = %d, want 2 (FK + view)", len(impacted))
	}

	types := map[string]bool{}
	for _, i := range impacted {
		types[i.ObjectType] = true
	}
	if !types["table"] || !types["view"] {
		t.Errorf("expected both table and view impacted, got %+v", impacted)
	}

	if impacted := d.AnalyzeImpact(json.RawMessage("garbage"), breaking); impacted != nil {
		t.Errorf("AnalyzeImpact with invalid metadata = %+v, want nil", impacted)
	}
}

func TestDiffer_FieldChangeComparison(t *testing.T) {
	t.Parallel()

	changes := []FieldChange{{Field: "a.b.type", Before: "int", After: "bigint"}}
	if !reflect.DeepEqual(changes, []FieldChange{{Field: "a.b.type", Before: "int", After: "bigint"}}) {
		t.Error("FieldChange comparison failed")
	}
}
