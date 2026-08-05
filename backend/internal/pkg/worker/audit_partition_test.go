package worker

import (
	"testing"
	"time"
)

func TestParsePartitionRange(t *testing.T) {
	start, end, ok := parsePartitionRange("FOR VALUES FROM ('2026-10-01 00:00:00+00') TO ('2027-01-01 00:00:00+00')")
	if !ok {
		t.Fatal("expected parse to succeed")
	}
	wantStart := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	wantEnd := time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC)
	if !start.Equal(wantStart) || !end.Equal(wantEnd) {
		t.Fatalf("got [%v, %v), want [%v, %v)", start, end, wantStart, wantEnd)
	}
}

func TestParsePartitionRangeUnparseable(t *testing.T) {
	cases := []string{
		"",
		"DEFAULT",
		"FOR VALUES FROM (MINVALUE) TO ('2026-10-01 00:00:00+00')",
		"FOR VALUES FROM ('garbage') TO ('2026-10-01 00:00:00+00')",
	}
	for _, c := range cases {
		if _, _, ok := parsePartitionRange(c); ok {
			t.Fatalf("expected %q to be unparseable", c)
		}
	}
}

func TestRangeOverlapSemantics(t *testing.T) {
	cases := []struct {
		name     string
		bound    string
		from, to time.Time
		want     bool
	}{
		{
			name:  "buffer partition covers next month",
			bound: "FOR VALUES FROM ('2026-10-01 00:00:00+00') TO ('2027-01-01 00:00:00+00')",
			from:  time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
			to:    time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC),
			want:  true,
		},
		{
			name:  "adjacent partition does not overlap",
			bound: "FOR VALUES FROM ('2026-10-01 00:00:00+00') TO ('2027-01-01 00:00:00+00')",
			from:  time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC),
			to:    time.Date(2027, 2, 1, 0, 0, 0, 0, time.UTC),
			want:  false,
		},
		{
			name:  "partition after buffer does not overlap",
			bound: "FOR VALUES FROM ('2026-07-01 00:00:00+00') TO ('2026-08-01 00:00:00+00')",
			from:  time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
			to:    time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC),
			want:  false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			start, end, ok := parsePartitionRange(c.bound)
			if !ok {
				t.Fatalf("unexpected parse failure for %q", c.bound)
			}
			got := start.Before(c.to) && end.After(c.from)
			if got != c.want {
				t.Fatalf("got %v, want %v", got, c.want)
			}
		})
	}
}
