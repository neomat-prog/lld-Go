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

}
