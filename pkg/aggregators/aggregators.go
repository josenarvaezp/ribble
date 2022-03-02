package aggregators

import (
	"encoding/json"
	"errors"
	"sync"
)

// Aggregator is an interface used to define new aggregators
type Aggregator interface {
	Reduce(messageBody *string) error
	UpdateOutput(intermediate interface{}, wg *sync.WaitGroup) error
}

// MapSum aggregates values from the same key by adding
// the values up
type MapSum map[string]int

// Reduce processes a message emmited by a mapper
func (ms MapSum) Reduce(messageBody *string) error {
	// unmarshall message body
	var res MessageInt
	body := []byte(*messageBody)
	err := json.Unmarshal(body, &res)
	if err != nil {
		return err
	}

	// process message
	currentKey := res.Key
	currentValue := res.Value

	// only process value if it is not empty
	// empty values are sent to keep the same number of events per batch
	if res.EmptyVal != true {
		ms[currentKey] = ms[currentKey] + currentValue
	}

	return nil
}

// UpdateOutput merges the outputMap with the intermediate map
func (ms MapSum) UpdateOutput(intermediateMap interface{}, wg *sync.WaitGroup) error {
	defer wg.Done()

	// cast intermediate map
	intermediateMapCast, ok := intermediateMap.(MapSum)
	if !ok {
		return errors.New("Error updating output")
	}

	// update output map values
	for k, v := range intermediateMapCast {
		ms[k] = ms[k] + v
	}

	return nil
}

// MapMax aggregates values from the same key
// by getting the max value of the given key
type MapMax map[string]int

// Reduce processes a message emmited by a mapper
func (mm MapMax) Reduce(messageBody *string) error {
	// unmarshall message body
	var res MessageInt
	body := []byte(*messageBody)
	err := json.Unmarshal(body, &res)
	if err != nil {
		return err
	}

	// process message
	currentKey := res.Key
	currentValue := res.Value

	// only process value if it is not empty
	// empty values are sent to keep the same number of events per batch
	if res.EmptyVal != true {
		previousMax, ok := mm[currentKey]
		if ok && previousMax < currentValue {
			// update new value
			mm[currentKey] = currentValue
		} else if !ok {
			// there was no previous value so we
			// add the new as the new max
			mm[currentKey] = currentValue
		}
	}

	return nil
}

// UpdateOutput merges the outputMap with the intermediate map
func (mm MapMax) UpdateOutput(intermediateMap interface{}, wg *sync.WaitGroup) error {
	defer wg.Done()

	// cast intermediate map
	intermediateMapCast, ok := intermediateMap.(MapMax)
	if !ok {
		return errors.New("Error updating output")
	}

	// update output map values
	for k, v := range intermediateMapCast {
		previousMax, ok := mm[k]
		if ok && previousMax < v {
			// update new value
			mm[k] = v
		} else if !ok {
			// there was no previous value so we
			// add the new as the new max
			mm[k] = v
		}
	}

	return nil
}

// MapMin aggregates values from the same key
// by getting the min value of the given key
type MapMin map[string]int

// Reduce processes a message emmited by a mapper
func (mm MapMin) Reduce(messageBody *string) error {
	// unmarshall message body
	var res MessageInt
	body := []byte(*messageBody)
	err := json.Unmarshal(body, &res)
	if err != nil {
		return err
	}

	// process message
	currentKey := res.Key
	currentValue := res.Value

	// only process value if it is not empty
	// empty values are sent to keep the same number of events per batch
	if res.EmptyVal != true {
		previousMin, ok := mm[currentKey]
		if ok && previousMin > currentValue {
			// update new value
			mm[currentKey] = currentValue
		} else if !ok {
			// there was no previous value so we
			// add the new as the new min
			mm[currentKey] = currentValue
		}
	}

	return nil
}

// UpdateOutput merges the outputMap with the intermediate map
func (mm MapMin) UpdateOutput(intermediateMap interface{}, wg *sync.WaitGroup) error {
	defer wg.Done()

	// cast intermediate map
	intermediateMapCast, ok := intermediateMap.(MapMin)
	if !ok {
		return errors.New("Error updating output")
	}

	// update output map values
	for k, v := range intermediateMapCast {
		previousMin, ok := mm[k]
		if ok && previousMin > v {
			// update new value as intermediate has
			// a lower value
			mm[k] = v
		} else if !ok {
			// there is no value in output map
			// so add value
			mm[k] = v
		}
	}

	return nil
}

// MessageInt represent a value emmited by a MapSum or a MapMax mapper
type MessageInt struct {
	Key      string `json:"key,omitempty"`
	Value    int    `json:"value,omitempty"`
	EmptyVal bool   `json:"empty,omitempty"`
}
