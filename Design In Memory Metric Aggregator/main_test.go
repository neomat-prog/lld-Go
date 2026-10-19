package main

import (
	"testing"
	"time"
)

func TestIngest(t *testing.T) {
	now := time.Date(2026, 10, 19, 12, 5, 30, 0, time.UTC)
	agg := NewAggregator(time.Hour, 10000, func() time.Time { return now })

	for _, v := range []float64{120, 90, 150} {
		p := Point{
			Name:      "latency",
			Tags:      map[string]string{"host": "a"},
			Value:     v,
			Timestamp: now,
		}
		if err := agg.Ingest(p); err != nil {
			t.Fatalf("ingest %v failed %v", v, err)
		}
	}
	key := seriesKey("latency", map[string]string{"host": "a"})
	window := now.Truncate(time.Minute).Unix()

	s, ok := agg.series[key]
	if !ok {
		t.Fatalf("series %q not found", key)
	}

	b, ok := s.Buckets[window]
	if !ok {
		t.Fatalf("bucket %d not found", window)
	}

	if b.Count != 3 {
		t.Errorf("Count: expected 3, got %d", b.Count)
	}
	if b.Sum != 360 {
		t.Errorf("Sum: expected 360, got %v", b.Sum)
	}
	if b.Min != 90 {
		t.Errorf("Min: expected 90, got %v", b.Min)
	}
	if b.Max != 150 {
		t.Errorf("Max: expected 150, got %v", b.Max)
	}
}
