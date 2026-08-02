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
	Type       string          `json:"type"`
	Name       string          `json:"name"`
	Definition json.RawMessage `json:"definition,omitempty"`
	Changes    []FieldChange   `json:"changes,omitempty"`
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

// BreakingChange represents a detected breaking or cautionary schema change.
type BreakingChange struct {
	ObjectType  string `json:"object_type"`
	ObjectName  string `json:"object_name"`
	Change      string `json:"change"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
}

// ImpactedObject represents a database object impacted by a breaking change.
type ImpactedObject struct {
	ObjectType string `json:"object_type"`
	ObjectName string `json:"object_name"`
	Dependency string `json:"dependency"`
	Impact     string `json:"impact"`
}

type viewInfo struct {
	Name       string `json:"name"`
	Schema     string `json:"schema"`
	Definition string `json:"definition"`
}

type impactMetadata struct {
	Tables []TableInfo `json:"tables"`
	Views  []viewInfo  `json:"views,omitempty"`
}

// DetectBreakingChanges analyzes a DiffResult and returns any breaking or cautionary changes found.
func (d *Differ) DetectBreakingChanges(diff *DiffResult) []BreakingChange {
	var changes []BreakingChange

	for _, obj := range diff.RemovedObjects {
		if obj.Type == "table" {
			changes = append(changes, BreakingChange{
				ObjectType:  "table",
				ObjectName:  obj.Name,
				Change:      "removed",
				Severity:    "breaking",
				Description: fmt.Sprintf("Table %s has been removed", obj.Name),
			})
		}
	}

	for _, obj := range diff.ModifiedObjects {
		for _, ch := range obj.Changes {
			parts := strings.SplitN(ch.Field, ".", 3)
			if len(parts) != 3 {
				continue
			}
			colName := parts[1]
			action := parts[2]
			fullColName := obj.Name + "." + colName

			switch action {
			case "removed":
				changes = append(changes, BreakingChange{
					ObjectType:  "column",
					ObjectName:  fullColName,
					Change:      "removed",
					Severity:    "breaking",
					Description: fmt.Sprintf("Column %s has been removed", fullColName),
				})
			case "type":
				changes = append(changes, BreakingChange{
					ObjectType:  "column",
					ObjectName:  fullColName,
					Change:      "type_changed",
					Severity:    "breaking",
					Description: fmt.Sprintf("Column %s type changed from %v to %v", fullColName, ch.Before, ch.After),
				})
			case "nullable":
				before, beforeOk := ch.Before.(bool)
				after, afterOk := ch.After.(bool)
				if beforeOk && afterOk {
					severity := "breaking"
					if !before && after {
						severity = "caution"
					}
					changes = append(changes, BreakingChange{
						ObjectType:  "column",
						ObjectName:  fullColName,
						Change:      "nullable_changed",
						Severity:    severity,
						Description: fmt.Sprintf("Column %s nullable changed from %v to %v", fullColName, before, after),
					})
				}
			}
		}
	}

	return changes
}

// AnalyzeImpact takes the version metadata and a list of breaking changes, and returns all
// database objects impacted by those changes (e.g. foreign keys referencing changed tables, views depending on them).
func (d *Differ) AnalyzeImpact(meta []byte, breakingChanges []BreakingChange) []ImpactedObject {
	var m impactMetadata
	if err := json.Unmarshal(meta, &m); err != nil {
		return nil
	}

	var impacted []ImpactedObject

	changedTables := make(map[string]bool)
	for _, bc := range breakingChanges {
		if bc.ObjectType == "table" {
			changedTables[bc.ObjectName] = true
		}
	}

	for _, table := range m.Tables {
		fullName := table.Schema + "." + table.Name
		for _, fk := range table.Constr.ForeignKeys {
			refFull := table.Schema + "." + fk.RefTable
			if changedTables[refFull] {
				impacted = append(impacted, ImpactedObject{
					ObjectType: "table",
					ObjectName: fullName,
					Dependency: fmt.Sprintf("foreign_key:%s", fk.Name),
					Impact:     fmt.Sprintf("Table %s has a foreign key %s referencing %s", fullName, fk.Name, refFull),
				})
			}
		}
	}

	for _, view := range m.Views {
		fullName := view.Schema + "." + view.Name
		for tbl := range changedTables {
			if strings.Contains(view.Definition, tbl) {
				impacted = append(impacted, ImpactedObject{
					ObjectType: "view",
					ObjectName: fullName,
					Dependency: fmt.Sprintf("view_definition:%s", tbl),
					Impact:     fmt.Sprintf("View %s references table %s", fullName, tbl),
				})
				break
			}
		}
	}

	return impacted
}
