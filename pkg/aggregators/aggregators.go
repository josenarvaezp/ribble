package aggregators

import (
	"errors"
	"fmt"
	"sync"
)

type AggregatorType int64

const (
	// Aggregator types
	InvalidAggregator AggregatorType = iota
	MapAggregatorType
	SumAggregatorType
	MaxAggregatorType
	MinAggregatorType
	AvgAggregatorType
)

// Aggregator is an interface used to define new aggregators
type Aggregator interface {
	Reduce(message *ReduceMessage) error
	UpdateOutput(intermediate interface{}, wg *sync.WaitGroup) error
	ToNum() float64
	Type() AggregatorType
}

// MapAggregator is used if the user needs to aggregate the
// values using different function
type MapAggregator map[string]Aggregator

func (ma MapAggregator) Type() AggregatorType {
	return MapAggregatorType
}

// Reduce reduces a specific element according to the
// message received from the mapper
func (ma MapAggregator) Reduce(message *ReduceMessage) error {
	if message.EmptyVal {
		// don't need to process the empty message
		return nil
	}

	_, ok := ma[message.Key]
	if !ok {
		// aggregator has not been initialized
		switch message.Type {
		case int64(SumAggregatorType):
			ma[message.Key] = InitSum(message.Value)
			return nil
		case int64(MaxAggregatorType):
			ma[message.Key] = InitMax(message.Value)
			return nil
		case int64(MinAggregatorType):
			ma[message.Key] = InitMin(message.Value)
			return nil
		case int64(AvgAggregatorType):
			ma[message.Key] = InitAvg(message.Value, message.Count)
			return nil
		default:
			errMessage := fmt.Sprintf("Invalid aggregator used, got: %d for value %f", message.Type, message.Value)
			return errors.New(errMessage)
		}
	}

	return ma[message.Key].Reduce(message)
}

// UpdateOutput updates all elements in the map accordingly.
// For example, if the element is Sum then the values get added and
// if it is Max then the element takes the greater value.
func (ma MapAggregator) UpdateOutput(intermediate interface{}, wg *sync.WaitGroup) error {
	defer wg.Done()

	// cast intermediate map
	intermediateValueCast, ok := intermediate.(MapAggregator)
	if !ok {
		return errors.New("Error updating output")
	}

	for key, aggregator := range intermediateValueCast {
		if _, ok := ma[key]; ok {
			ma[key].UpdateOutput(aggregator, nil)
		} else {
			ma[key] = aggregator
		}
	}

	return nil
}

// ToNum is not implemented for MapAggregator
func (ma MapAggregator) ToNum() float64 {
	return -1
}

// AddSum is a helper function the user can use to add a
// Sum value to the aggregator map
func (ma MapAggregator) AddSum(key string, value float64) error {
	currentSum, ok := ma[key]
	if !ok {
		sum := Sum(value)
		ma[key] = &sum
	} else {
		// cast intermediate map
		castSum, ok := currentSum.(*Sum)
		if !ok {
			return errors.New("Mixed aggregators used")
		}

		sum := Sum(castSum.ToNum() + currentSum.ToNum())
		ma[key] = &sum
	}

	return nil
}

// AddMax is a helper function the user can use to add a
// Max value to the aggregator map
func (ma MapAggregator) AddMax(key string, value float64) error {
	currentMax, ok := ma[key]
	if !ok {
		max := Max(value)
		ma[key] = &max
	} else {
		// cast intermediate map
		castMax, ok := currentMax.(*Max)
		if !ok {
			return errors.New("Mixed aggregators used")
		}

		if castMax.ToNum() < value {
			// update new max
			max := Max(value)
			ma[key] = &max
		}
	}

	return nil
}

// AddMin is a helper function the user can use to add a
// Min value to the aggregator map
func (ma MapAggregator) AddMin(key string, value float64) error {
	currentMin, ok := ma[key]
	if !ok {
		min := Min(value)
		ma[key] = &min
	} else {
		// cast intermediate map
		castMax, ok := currentMin.(*Min)
		if !ok {
			return errors.New("Mixed aggregators used")
		}

		if castMax.ToNum() > value {
			// update new min
			min := Min(value)
			ma[key] = &min
		}
	}

	return nil
}

// AddAvg is a helper function the user can use to add a
// Avg value to the aggregator map
func (ma MapAggregator) AddAvg(key string, value float64) error {
	currentAvg, ok := ma[key]
	if !ok {
		ma[key] = &Avg{
			Sum:   value,
			Count: 1,
		}
	} else {
		// cast intermediate map
		castAvg, ok := currentAvg.(*Avg)
		if !ok {
			return errors.New("Mixed aggregators used")
		}

		ma[key] = &Avg{
			Sum:   castAvg.Sum + value,
			Count: castAvg.Count + 1,
		}
	}

	return nil
}

// -------------------
// SUM AGGREGATOR
// -------------------

// Sum aggregates values emitted by adding them up
type Sum float64

// InitSum initializes a Sum value to 0
func InitSum(value float64) *Sum {
	sum := Sum(value)
	return &sum
}

// Add updates the sum by adding the new value
func (s *Sum) Add(value float64) {
	newSum := Sum(s.ToNum() + value)
	s = &newSum
}

// ToNum converts the Sum value to a float
func (s *Sum) ToNum() float64 {
	return float64(*s)
}

func (ma *Sum) Type() AggregatorType {
	return SumAggregatorType
}

// Reduce aggregates values emitted by adding them up
func (s *Sum) Reduce(message *ReduceMessage) error {
	newVal := Sum(s.ToNum() + message.Value)
	s = &newVal

	return nil
}

// UpdateOutput merges the previous Sum value by adding the new intermediate value
func (s *Sum) UpdateOutput(intermediateValue interface{}, wg *sync.WaitGroup) error {
	// cast intermediate map
	intermediateValueCast, ok := intermediateValue.(*Sum)
	if !ok {
		return errors.New("Error updating output")
	}

	// update output map values
	newVal := Sum(s.ToNum() + intermediateValueCast.ToNum())
	s = &newVal

	return nil
}

// -------------------
// MAX AGGREGATOR
// -------------------

// MapMax aggregates values from the same key
// by getting the max value of the given key
type Max float64

// InitMax initializes a Max value to the minimum
// value a float can take
func InitMax(value float64) *Max {
	max := Max(value)
	return &max
}

func (m *Max) ToNum() float64 {
	return float64(*m)
}

func (ma *Max) Type() AggregatorType {
	return MaxAggregatorType
}

// Reduce processes a message emmited by a mapper
func (m *Max) Reduce(message *ReduceMessage) error {
	if m.ToNum() < message.Value {
		// update new max
		newMax := Max(message.Value)
		m = &newMax
	}

	return nil
}

// UpdateOutput merges the outputMap with the intermediate map
func (m *Max) UpdateOutput(intermediateValue interface{}, wg *sync.WaitGroup) error {
	// cast intermediate map
	intermediateValueCast, ok := intermediateValue.(*Max)
	if !ok {
		return errors.New("Error updating output")
	}

	// update max
	if m.ToNum() < intermediateValueCast.ToNum() {
		// update new max
		m = intermediateValueCast
	}

	return nil
}

// -------------------
// MIN AGGREGATOR
// -------------------

// Min aggregates values from the same key
// by getting the max value of the given key
type Min float64

// InitMin initializes a Max value to the minimum
// value a float can take
func InitMin(value float64) *Min {
	min := Min(value)
	return &min
}

func (m *Min) ToNum() float64 {
	return float64(*m)
}

func (ma *Min) Type() AggregatorType {
	return MinAggregatorType
}

// Reduce processes a message emmited by a mapper
func (m *Min) Reduce(message *ReduceMessage) error {
	if m.ToNum() > message.Value {
		// update new max
		newMin := Min(message.Value)
		m = &newMin
	}

	return nil
}

// UpdateOutput merges the outputMap with the intermediate map
func (m *Min) UpdateOutput(intermediateValue interface{}, wg *sync.WaitGroup) error {
	// cast intermediate map
	intermediateValueCast, ok := intermediateValue.(*Min)
	if !ok {
		return errors.New("Error updating output")
	}

	// update max
	if m.ToNum() > intermediateValueCast.ToNum() {
		// update new min
		m = intermediateValueCast
	}

	return nil
}

// -------------------
// AVG AGGREGATOR
// -------------------

// Avg aggregates values emitted by performing their average
type Avg struct {
	Avg   float64 `json:",string,omitempty"`
	Sum   float64 `json:"-"`
	Count int     `json:"-"`
}

func (a *Avg) GetSum() float64 {
	return a.Sum
}

func (a *Avg) GetCount() int {
	return a.Count
}

// InitAvg initializes a Avg value to 0
func InitAvg(value float64, count int) *Avg {
	return &Avg{
		Avg:   value / float64(count),
		Sum:   value,
		Count: count,
	}
}

// ToNum converts the Avg value to a float
func (a *Avg) ToNum() float64 {
	return a.Avg
}

// PerformAvg performs the average
func (a *Avg) PerformAvg() {
	a.Avg = a.Sum / float64(a.Count)
}

func (ma *Avg) Type() AggregatorType {
	return AvgAggregatorType
}

// Reduce aggregates values emitted by adding them up
// it is important to notice that the average computation
// is done in the reducer lambda
func (a *Avg) Reduce(message *ReduceMessage) error {
	a.Sum = a.Sum + message.Value
	a.Count = a.Count + message.Count
	a.PerformAvg()

	return nil
}

// UpdateOutput merges the previous Avg value by adding the new intermediate value
func (a *Avg) UpdateOutput(intermediateValue interface{}, wg *sync.WaitGroup) error {
	// cast intermediate map
	intermediateValueCast, ok := intermediateValue.(*Avg)
	if !ok {
		return errors.New("Error updating output")
	}

	// update output map values
	a.Sum = a.Sum + intermediateValueCast.Sum
	a.Count = a.Count + intermediateValueCast.Count

	a.Avg = a.Sum / float64(a.Count)

	return nil
}

// ReduceMessage represent a value emmited
type ReduceMessage struct {
	Key      string  `json:"key,omitempty"`
	Value    float64 `json:"value,omitempty"`
	Count    int     `json:"count,omitempty"`
	Type     int64   `json:"type,omitempty"`
	EmptyVal bool    `json:"empty,omitempty"`
}
