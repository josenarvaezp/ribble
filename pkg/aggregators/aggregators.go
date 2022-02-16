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
type MapSum map[string]Sum

// Reduce processes a message emmited by a mapper
func (ms MapSum) Reduce(messageBody *string) error {
	// unmarshall message body
	var res MessageSum
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

type Sum int

func (c Sum) Int() int {
	return int(c)
}

func (s *Sum) Reduce(messageBody *string) error {
	// unmarshall message body
	var res MessageSum
	body := []byte(*messageBody)
	err := json.Unmarshal(body, &res)
	if err != nil {
		return err
	}

	// process message
	currentValue := res.Value

	// only process value if it is not empty
	// empty values are sent to keep the same number of events per batch
	if res.EmptyVal != true {
		newVal := *s + currentValue
		s = &newVal
	}

	return nil
}

// MessageSum represent a value emmited by a MapSum mapper
type MessageSum struct {
	Key      string `json:"key,omitempty"`
	Value    Sum    `json:"value,omitempty"`
	EmptyVal bool   `json:"empty,omitempty"`
}
