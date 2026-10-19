package main

import (
	"errors"
	"math"
	"sort"
	"time"
)

// Data Point
type Point struct {
	Name      string
	Tags      map[string]string
	Value     float64
	Timestamp time.Time
}

type Bucket struct {
	Count int
	Sum   float64
	Min   float64
	Max   float64
}

type Series struct {
	Name    string
	Tags    map[string]string
	Buckets map[int64]*Bucket
}

type Clock func() time.Time

var (
	ErrInvalidPoint = errors.New("invalid point")
)

// Point -> ingest -> turn name + tags into single map key, so that tag order doesn't matter
type Aggregator struct {
	series    map[string]*Series // ("latency{host=a,region=eu}")
	retention time.Duration
	maxSeries int
	clock     Clock
}

// Export every flush interval (eg. 10s)
// type Export interface {
//  Export(buckets []SeriesBucket) error
// }

func NewAggregator(retention time.Duration, maxSeries int, clock Clock) *Aggregator {
	return &Aggregator{
		series:    make(map[string]*Series),
		retention: retention,
		maxSeries: maxSeries,
		clock:     clock,
	}
}

func seriesKey(name string, tags map[string]string) string {
	keys := []string{}
	for key := range tags {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// { "host" "a" "region" "eu"}

	result := name + "{"
	for i, k := range keys {
		if i > 0 {
			result += ","
		}
		result += k + "=" + tags[k]
	}
	result += "}"
	return result
}

func (a *Aggregator) validate(p Point) error {
	now := a.clock()

	if p.Name == "" {
		return ErrInvalidPoint
	}

	if math.IsNaN(p.Value) || math.IsInf(p.Value, 0) {
		return ErrInvalidPoint
	}

	if p.Timestamp.After(now.Add(time.Minute)) {
		return ErrInvalidPoint
	}

	if p.Timestamp.Before(now.Add(-a.retention)) {
		return ErrInvalidPoint
	}

	return nil
}

func (a *Aggregator) Ingest(p Point) error {
	if err := a.validate(p); err != nil {
		return err
	}

	// build the key with the series key
	key := seriesKey(p.Name, p.Tags)

	s, ok := a.series[key]
	if !ok {
		s = &Series{
			Name:    p.Name,
			Tags:    p.Tags,
			Buckets: make(map[int64]*Bucket),
		}
		a.series[key] = s
	}

	window := p.Timestamp.Truncate(time.Minute).Unix()
	b, ok := s.Buckets[window]
	if !ok {
		b = &Bucket{}
		s.Buckets[window] = b
	}

	if b.Count == 0 {
		b.Min = p.Value
		b.Max = p.Value
	} else {
		if p.Value < b.Min {
			b.Min = p.Value
		}
		if p.Value > b.Max {
			b.Max = p.Value
		}
	}
	b.Count++
	b.Sum += p.Value

	return nil
}
