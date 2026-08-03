// Package pagination encodes and decodes stable, tie-free cursor values for
// keyset pagination. A cursor bundles the sort timestamp with the row ID so
// rows sharing an identical timestamp can never be skipped or duplicated
// between pages.
package pagination

import (
	"strings"
	"time"
)

const separator = "|"

// Encode builds a cursor from a sort timestamp and a row ID.
func Encode(ts time.Time, id string) string {
	return ts.UTC().Format(time.RFC3339Nano) + separator + id
}

// Decode splits a cursor produced by Encode back into its parts. It returns
// ok=false for empty or malformed cursors.
func Decode(cursor string) (time.Time, string, bool) {
	if cursor == "" {
		return time.Time{}, "", false
	}
	ts, id, found := strings.Cut(cursor, separator)
	if !found || id == "" {
		return time.Time{}, "", false
	}
	t, err := time.Parse(time.RFC3339Nano, ts)
	if err != nil {
		return time.Time{}, "", false
	}
	return t, id, true
}

// DecodeTime is a convenience for callers that only need the timestamp half.
func DecodeTime(cursor string) (time.Time, bool) {
	t, _, ok := Decode(cursor)
	return t, ok
}

// DecodeID is a convenience for callers that only need the ID half.
func DecodeID(cursor string) (string, bool) {
	_, id, ok := Decode(cursor)
	return id, ok
}

// EncodeTuple joins arbitrary sort-key parts (timestamp, name, type, ...)
// into a single cursor. Parts must not contain the separator character.
func EncodeTuple(parts ...string) string {
	return strings.Join(parts, separator)
}

// DecodeTuple splits a cursor produced by EncodeTuple back into its parts.
func DecodeTuple(cursor string) ([]string, bool) {
	if cursor == "" {
		return nil, false
	}
	parts := strings.Split(cursor, separator)
	if len(parts) < 2 {
		return nil, false
	}
	for _, p := range parts {
		if p == "" {
			return nil, false
		}
	}
	return parts, true
}
