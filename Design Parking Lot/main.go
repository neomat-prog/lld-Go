package main

import (
	"math"
	"sync"
	"time"
)

type Size string
type VType string

const (
	Motorcycle VType = "motocycle"
	Car        VType = "car"
	Truck      VType = "truck"
)

const (
	Large  Size = "large"
	Medium Size = "medium"
	Small  Size = "small"
)

type Vehicle struct {
	number string
	vtype  VType
}

type ParkingSpot struct {
	id       int
	size     Size
	parkedBy *Vehicle
}

func (s *ParkingSpot) isAvailable() bool {
	return s.parkedBy == nil
}

func CanFit(v VType, size Size) bool {
	switch v {
	case Motorcycle:
		return size == Large || size == Medium || size == Small
	case Car:
		return size == Medium || size == Large
	case Truck:
		return size == Large
	default:
		return false
	}
}

type ParkingFloor struct {
	number int
	spots  []*ParkingSpot
}

func NewParkingFloor(number int, sizes []Size) *ParkingFloor {
	floor := &ParkingFloor{
		number: number,
	}

	for i, size := range sizes {
		floor.spots = append(floor.spots, &ParkingSpot{
			id:   i + 1,
			size: size,
		})
	}

	return floor
}

type Ticket struct {
	id        int
	vehicle   *Vehicle
	floor     int
	spotID    int
	entryTime time.Time
	exitTime  *time.Time
	fee       float64
}

type TicketBooth struct {
	pricing PricingStrategy
}

type PricingStrategy interface {
	CalculateFee(ticket *Ticket, exitTime time.Time) float64
}

type FlatRatePricing struct {
	Amount float64
}

func (p FlatRatePricing) CalculateFee(ticket *Ticket, exitTime time.Time) float64 {
	return p.Amount
}

type VehicleTypePricing struct {
	HourlyRates map[VType]float64
}

func (p VehicleTypePricing) CalculateFee(ticket *Ticket, exitTime time.Time) float64 {
	dur := exitTime.Sub(ticket.entryTime).Hours()
	hrs := math.Max(1, math.Ceil(dur))

	rate := p.HourlyRates[ticket.vehicle.vtype]
	return hrs * rate
}

func (b *TicketBooth) CalculateFee(ticket *Ticket, exitTime time.Time) float64 {
	return b.pricing.CalculateFee(ticket, exitTime)
}

type ParkingLot struct {
	floors      []*ParkingFloor
	ticketBooth *TicketBooth

	mu           sync.Mutex
	nextTicketID int
	active       map[int]*Ticket
}

func NewParkingLot(pricing PricingStrategy) *ParkingLot {
	return &ParkingLot{
		ticketBooth: &TicketBooth{pricing: pricing},
		active:      make(map[int]*Ticket),
	}
}

func (p *ParkingLot) AddFloor(floor *ParkingFloor) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.floors = append(p.floors, floor)
}

// find the smallest available spot
func (p *ParkingLot) findSpot(vType VType) (*ParkingFloor, *ParkingSpot) {
	sizes := []Size{Small, Medium, Large}

	for _, size := range sizes {
		if !CanFit(vType, size) {
			continue
		}

		for _, floor := range p.floors {
			for _, spot := range floor.spots {
				if spot.size == size && spot.isAvailable() {
					return floor, spot
				}
			}
		}
	}

	return nil, nil
}
