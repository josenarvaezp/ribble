package aggregators

import (
	"errors"
	"sync"
)

// Aggregator is an interface used to define new aggregators
type Aggregator interface {
	Reduce(message *ReduceMessage) error
	UpdateOutput(intermediate interface{}, wg *sync.WaitGroup) error
	ToNum() float64
}

// MapAggregator is used if the user needs to aggregate the
// values using different function
type MapAggregator map[string]Aggregator

// Reduce reduces a specific element according to the
// message received from the mapper
func (ma MapAggregator) Reduce(message *ReduceMessage) error {
	aggregator, ok := ma[message.Key]
	if !ok {
		// aggregator has not been initialized
		switch message.Type {
		case 2:
			aggregator = InitSum()
		default:
			return errors.New("Invalid aggregator used")
		}
	}

	return aggregator.Reduce(message)
}

// UpdateOutput updates all elements in the map accordingly.
// For example, if the element is Sum then the values get added and
// if it is Max then the element takes the greater value.
func (ma MapAggregator) UpdateOutput(intermediate interface{}, wg *sync.WaitGroup) error {
	// cast intermediate map
	intermediateValueCast, ok := intermediate.(MapAggregator)
	if !ok {
		return errors.New("Error updating output")
	}

	for key, aggregator := range intermediateValueCast {
		ma[key].UpdateOutput(aggregator, wg)
	}

	return nil
}

// ToNum is not implemented for MapAggregator
func (ma MapAggregator) ToNum() float64 {
	return -1
}

// AddSum is a helper function the user can use to add a
// Sum value to the aggregator map
func (ma MapAggregator) AddSum(key string, value int) {
	ma[key] = Sum(value)
}

// Sum aggregates values emitted by adding them up
type Sum float64

// InitSum initializes a Sum value to 0
func InitSum() Sum {
	return Sum(0)
}

// Add updates the sum by adding the new value
func (c Sum) Add(value float64) {
	c = c + Sum(value)
}

// ToNum converts the Sum value to a float
func (a Sum) ToNum() float64 {
	return 0
}

// Reduce aggregates values emitted by adding them up
func (s Sum) Reduce(message *ReduceMessage) error {
	// process message
	currentValue := message.Value

	// only process value if it is not empty
	// empty values are sent to keep the same number of events per batch
	if message.EmptyVal != true {
		newVal := Sum(s.ToNum() + currentValue)
		s = newVal
	}

	return nil
}

// UpdateOutput merges the previous Sum value by adding the new intermediate value
func (s Sum) UpdateOutput(intermediateValue interface{}, wg *sync.WaitGroup) error {
	defer wg.Done()

	// cast intermediate map
	intermediateValueCast, ok := intermediateValue.(*Sum)
	if !ok {
		return errors.New("Error updating output")
	}

	// update output map values
	newVal := Sum(s.ToNum() + intermediateValueCast.ToNum())
	s = newVal

	return nil
}

// ReduceMessage represent a value emmited
type ReduceMessage struct {
	Key      string  `json:"key,omitempty"`
	Value    float64 `json:"value,omitempty"`
	Type     int64   `json:"type,omitempty"`
	EmptyVal bool    `json:"empty,omitempty"`
}

// func (ma MapAggregator) AddMax(key string, value int) {
// 	ma[key] = Max(value)
// }

// func (ma MapAggregator) AddAvg(key string, value int) {
// 	ma[key] = Sum(value)
// }

// // MapSum aggregates values from the same key by adding
// // the values up
// type MapSum map[string]int

// func (a MapSum) ToNum() float64 {
// 	return 0
// }

// // Reduce processes a message emmited by a mapper
// func (ms MapSum) Reduce(messageBody *string) error {
// 	// unmarshall message body
// 	var res MessageInt
// 	body := []byte(*messageBody)
// 	err := json.Unmarshal(body, &res)
// 	if err != nil {
// 		return err
// 	}

// 	// process message
// 	currentKey := res.Key
// 	currentValue := res.Value

// 	// only process value if it is not empty
// 	// empty values are sent to keep the same number of events per batch
// 	if res.EmptyVal != true {
// 		ms[currentKey] = ms[currentKey] + currentValue
// 	}

// 	return nil
// }

// // UpdateOutput merges the outputMap with the intermediate map
// func (ms MapSum) UpdateOutput(intermediateMap interface{}, wg *sync.WaitGroup) error {
// 	defer wg.Done()

// 	// cast intermediate map
// 	intermediateMapCast, ok := intermediateMap.(MapSum)
// 	if !ok {
// 		return errors.New("Error updating output")
// 	}

// 	// update output map values
// 	for k, v := range intermediateMapCast {
// 		ms[k] = ms[k] + v
// 	}

// 	return nil
// }

// // MapMax aggregates values from the same key
// // by getting the max value of the given key
// type MapMax map[string]int

// func (a MapMax) ToNum() float64 {
// 	return 0
// }

// // Reduce processes a message emmited by a mapper
// func (mm MapMax) Reduce(messageBody *string) error {
// 	// unmarshall message body
// 	var res MessageInt
// 	body := []byte(*messageBody)
// 	err := json.Unmarshal(body, &res)
// 	if err != nil {
// 		return err
// 	}

// 	// process message
// 	currentKey := res.Key
// 	currentValue := res.Value

// 	// only process value if it is not empty
// 	// empty values are sent to keep the same number of events per batch
// 	if res.EmptyVal != true {
// 		previousMax, ok := mm[currentKey]
// 		if ok && previousMax < currentValue {
// 			// update new value
// 			mm[currentKey] = currentValue
// 		} else if !ok {
// 			// there was no previous value so we
// 			// add the new as the new max
// 			mm[currentKey] = currentValue
// 		}
// 	}

// 	return nil
// }

// // UpdateOutput merges the outputMap with the intermediate map
// func (mm MapMax) UpdateOutput(intermediateMap interface{}, wg *sync.WaitGroup) error {
// 	defer wg.Done()

// 	// cast intermediate map
// 	intermediateMapCast, ok := intermediateMap.(MapMax)
// 	if !ok {
// 		return errors.New("Error updating output")
// 	}

// 	// update output map values
// 	for k, v := range intermediateMapCast {
// 		previousMax, ok := mm[k]
// 		if ok && previousMax < v {
// 			// update new value
// 			mm[k] = v
// 		} else if !ok {
// 			// there was no previous value so we
// 			// add the new as the new max
// 			mm[k] = v
// 		}
// 	}

// 	return nil
// }

// // MapMin aggregates values from the same key
// // by getting the min value of the given key
// type MapMin map[string]int

// func (a MapMin) ToNum() float64 {
// 	return 0
// }

// // Reduce processes a message emmited by a mapper
// func (mm MapMin) Reduce(messageBody *string) error {
// 	// unmarshall message body
// 	var res MessageInt
// 	body := []byte(*messageBody)
// 	err := json.Unmarshal(body, &res)
// 	if err != nil {
// 		return err
// 	}

// 	// process message
// 	currentKey := res.Key
// 	currentValue := res.Value

// 	// only process value if it is not empty
// 	// empty values are sent to keep the same number of events per batch
// 	if res.EmptyVal != true {
// 		previousMin, ok := mm[currentKey]
// 		if ok && previousMin > currentValue {
// 			// update new value
// 			mm[currentKey] = currentValue
// 		} else if !ok {
// 			// there was no previous value so we
// 			// add the new as the new min
// 			mm[currentKey] = currentValue
// 		}
// 	}

// 	return nil
// }

// // UpdateOutput merges the outputMap with the intermediate map
// func (mm MapMin) UpdateOutput(intermediateMap interface{}, wg *sync.WaitGroup) error {
// 	defer wg.Done()

// 	// cast intermediate map
// 	intermediateMapCast, ok := intermediateMap.(MapMin)
// 	if !ok {
// 		return errors.New("Error updating output")
// 	}

// 	// update output map values
// 	for k, v := range intermediateMapCast {
// 		previousMin, ok := mm[k]
// 		if ok && previousMin > v {
// 			// update new value as intermediate has
// 			// a lower value
// 			mm[k] = v
// 		} else if !ok {
// 			// there is no value in output map
// 			// so add value
// 			mm[k] = v
// 		}
// 	}

// 	return nil
// }

// type Avg int

// func (a Avg) Reduce(messageBody *string) error {
// 	return nil
// }

// func (a Avg) UpdateOutput(intermediateValue interface{}, wg *sync.WaitGroup) error {
// 	return nil
// }

// func (a Avg) ToNum() float64 {
// 	return 0
// }

// type Max int

// func (m Max) Reduce(messageBody *string) error {
// 	return nil
// }

// func (m Max) UpdateOutput(intermediateValue interface{}, wg *sync.WaitGroup) error {
// 	return nil
// }

// func (a Max) ToNum() float64 {
// 	return 0
// }
