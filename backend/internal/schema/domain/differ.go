package domain

import (
	"encoding/json"
	"fmt"
	"strings"
)

type DiffResult struct {
	AddedObjects    []DiffObject `json:"added_objects"`
	RemovedObjects  []DiffObject `json:"removed_objects"`
	ModifiedObjects []DiffObject `json:"modified_objects"`
}

type DiffObject struct {
	Type       string         `json:"type"`
	Name       string         `json:"name"`
	Definition json.RawMessage `json:"definition,omitempty"`
	Changes    []FieldChange  `json:"changes,omitempty"`
}

type FieldChange struct {
	Field  string      `json:"field"`
	Before interface{} `json:"before"`
	After  interface{} `json:"after"`
}

type Differ struct{}

func NewDiffer() *Differ {
	return &Differ{}
}

func (d *Differ) Diff(a, b json.RawMessage) (*DiffResult, error) {
	var metaA, metaB SchemaMetadata
	if err := json.Unmarshal(a, &metaA); err != nil {
		return nil, fmt.Errorf("unmarshaling version A: %w", err)
	}
	if err := json.Unmarshal(b, &metaB); err != nil {
		return nil, fmt.Errorf("unmarshaling version B: %w", err)
	}

	result := &DiffResult{}

	tableMapA := make(map[string]TableInfo)
	for _, t := range metaA.Tables {
		key := t.Schema + "." + t.Name
		tableMapA[key] = t
	}

	tableMapB := make(map[string]TableInfo)
	for _, t := range metaB.Tables {
		key := t.Schema + "." + t.Name
		tableMapB[key] = t
	}

	for key, tb := range tableMapB {
		if _, exists := tableMapA[key]; !exists {
			def, _ := json.Marshal(tb)
			result.AddedObjects = append(result.AddedObjects, DiffObject{
				Type:       "table",
				Name:       key,
				Definition: def,
			})
		}
	}

	for key, tb := range tableMapA {
		if _, exists := tableMapB[key]; !exists {
			def, _ := json.Marshal(tb)
			result.RemovedObjects = append(result.RemovedObjects, DiffObject{
				Type:       "table",
				Name:       key,
				Definition: def,
			})
		}
	}

	for key := range tableMapA {
		tbB, exists := tableMapB[key]
		if !exists {
			continue
		}
		tbA := tableMapA[key]

		changes := d.diffColumns(tbA.Columns, tbB.Columns, tbA.Name)
		changes = append(changes, d.diffIndexes(tbA.Indexes, tbB.Indexes)...)

		if len(changes) > 0 {
			result.ModifiedObjects = append(result.ModifiedObjects, DiffObject{
				Type:    "table",
				Name:    key,
				Changes: changes,
			})
		}
	}

	// Diff enums
	enumMapA := make(map[string]EnumInfo)
	for _, e := range metaA.Enums {
		enumMapA[e.Name] = e
	}
	enumMapB := make(map[string]EnumInfo)
	for _, e := range metaB.Enums {
		enumMapB[e.Name] = e
	}

	for name, eb := range enumMapB {
		if _, exists := enumMapA[name]; !exists {
			def, _ := json.Marshal(eb)
			result.AddedObjects = append(result.AddedObjects, DiffObject{
				Type: "enum", Name: name, Definition: def,
			})
		}
	}
	for name, ea := range enumMapA {
		if _, exists := enumMapB[name]; !exists {
			def, _ := json.Marshal(ea)
			result.RemovedObjects = append(result.RemovedObjects, DiffObject{
				Type: "enum", Name: name, Definition: def,
			})
		}
	}

	return result, nil
}

func (d *Differ) diffColumns(a, b []ColumnInfo, tableName string) []FieldChange {
	var changes []FieldChange

	colMapA := make(map[string]ColumnInfo)
	for _, c := range a {
		colMapA[c.Name] = c
	}
	colMapB := make(map[string]ColumnInfo)
	for _, c := range b {
		colMapB[c.Name] = c
	}

	for name, cb := range colMapB {
		if _, exists := colMapA[name]; !exists {
			changes = append(changes, FieldChange{
				Field:  fmt.Sprintf("%s.%s.added", tableName, name),
				Before: nil,
				After:  cb.DataType,
			})
		}
	}

	for name, ca := range colMapA {
		cb, exists := colMapB[name]
		if !exists {
			changes = append(changes, FieldChange{
				Field:  fmt.Sprintf("%s.%s.removed", tableName, name),
				Before: ca.DataType,
				After:  nil,
			})
			continue
		}
		if ca.DataType != cb.DataType {
			changes = append(changes, FieldChange{
				Field:  fmt.Sprintf("%s.%s.type", tableName, name),
				Before: ca.DataType,
				After:  cb.DataType,
			})
		}
		if ca.IsNullable != cb.IsNullable {
			changes = append(changes, FieldChange{
				Field:  fmt.Sprintf("%s.%s.nullable", tableName, name),
				Before: ca.IsNullable,
				After:  cb.IsNullable,
			})
		}
	}

	return changes
}

func (d *Differ) diffIndexes(a, b []IndexInfo) []FieldChange {
	var changes []FieldChange
	idxMapA := make(map[string]IndexInfo)
	for _, idx := range a {
		idxMapA[idx.Name] = idx
	}
	idxMapB := make(map[string]IndexInfo)
	for _, idx := range b {
		idxMapB[idx.Name] = idx
	}

	for name, ib := range idxMapB {
		if _, exists := idxMapA[name]; !exists {
			changes = append(changes, FieldChange{
				Field: fmt.Sprintf("index.%s.added", name), Before: nil, After: strings.Join(ib.Columns, ","),
			})
		}
	}
	for name, ia := range idxMapA {
		if _, exists := idxMapB[name]; !exists {
			changes = append(changes, FieldChange{
				Field: fmt.Sprintf("index.%s.removed", name), Before: strings.Join(ia.Columns, ","), After: nil,
			})
		}
	}

	return changes
}
