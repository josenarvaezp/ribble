package aggregators

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// this function reduces a set of values in an empty aggregator map
func Test_MapAggregatorReduceEmpty_HappyPath(t *testing.T) {
	// when reduced should be 1
	sumReduceMessage := &ReduceMessage{
		Key:      "a sum key",
		Value:    1,
		Type:     int64(SumAggregatorType),
		EmptyVal: false,
	}

	// when reduced should be 2
	maxReduceMessage := &ReduceMessage{
		Key:      "a max key",
		Value:    2,
		Type:     int64(MaxAggregatorType),
		EmptyVal: false,
	}

	// when reduced should be 3
	minReduceMessage := &ReduceMessage{
		Key:      "a min key",
		Value:    3,
		Type:     int64(MinAggregatorType),
		EmptyVal: false,
	}

	// when reduced should be 2
	avgReduceMessage := &ReduceMessage{
		Key:      "an avg key",
		Value:    4,
		Type:     int64(AvgAggregatorType),
		Count:    2,
		EmptyVal: false,
	}

	aggregatorMap := NewMap()

	// reduce messages where the key does not exist
	aggregatorMap.Reduce(sumReduceMessage)
	aggregatorMap.Reduce(maxReduceMessage)
	aggregatorMap.Reduce(minReduceMessage)
	aggregatorMap.Reduce(avgReduceMessage)

	// asser aggregated values
	assert.Equal(t, float64(1), aggregatorMap["a sum key"].ToNum())
	assert.Equal(t, float64(2), aggregatorMap["a max key"].ToNum())
	assert.Equal(t, float64(3), aggregatorMap["a min key"].ToNum())
	assert.Equal(t, float64(2), aggregatorMap["an avg key"].ToNum())

	// asser type of aggregator
	assert.Equal(t, SumAggregatorType, aggregatorMap["a sum key"].Type())
	assert.Equal(t, MaxAggregatorType, aggregatorMap["a max key"].Type())
	assert.Equal(t, MinAggregatorType, aggregatorMap["a min key"].Type())
	assert.Equal(t, AvgAggregatorType, aggregatorMap["an avg key"].Type())
}

// this function reduces a set of values in a map with existing keys
func Test_MapAggregatorReduceNotEmpty_HappyPath(t *testing.T) {
	// when reduced should be 3
	sumReduceMessage := &ReduceMessage{
		Key:      "a sum key",
		Value:    1,
		Type:     int64(SumAggregatorType),
		EmptyVal: false,
	}

	// when reduced should be 2
	maxReduceMessage := &ReduceMessage{
		Key:      "a max key",
		Value:    2,
		Type:     int64(MaxAggregatorType),
		EmptyVal: false,
	}

	// when reduced should be 3
	minReduceMessage := &ReduceMessage{
		Key:      "a min key",
		Value:    3,
		Type:     int64(MinAggregatorType),
		EmptyVal: false,
	}

	// when reduced should be 3
	avgReduceMessage := &ReduceMessage{
		Key:      "an avg key",
		Value:    4,
		Type:     int64(AvgAggregatorType),
		Count:    2,
		EmptyVal: false,
	}

	aggregatorMap := NewMap()
	aggregatorMap.AddSum("a sum key", 2)
	aggregatorMap.AddMax("a max key", 1)
	aggregatorMap.AddMin("a min key", 90)
	aggregatorMap.AddAvg("an avg key", 5)

	// reduce messages where the key does not exist
	aggregatorMap.Reduce(sumReduceMessage)
	aggregatorMap.Reduce(maxReduceMessage)
	aggregatorMap.Reduce(minReduceMessage)
	aggregatorMap.Reduce(avgReduceMessage)

	// assert aggregated values
	assert.Equal(t, float64(3), aggregatorMap["a sum key"].ToNum())
	assert.Equal(t, float64(2), aggregatorMap["a max key"].ToNum())
	assert.Equal(t, float64(3), aggregatorMap["a min key"].ToNum())
	assert.Equal(t, float64(3), aggregatorMap["an avg key"].ToNum())

	// asser type of aggregator
	assert.Equal(t, SumAggregatorType, aggregatorMap["a sum key"].Type())
	assert.Equal(t, MaxAggregatorType, aggregatorMap["a max key"].Type())
	assert.Equal(t, MinAggregatorType, aggregatorMap["a min key"].Type())
	assert.Equal(t, AvgAggregatorType, aggregatorMap["an avg key"].Type())
}

func Test_MapAggregatorAdd_HappyPath(t *testing.T) {
	aggregatorMap := NewMap()
	err := aggregatorMap.AddSum("a sum key", 2)
	assert.Nil(t, err)
	err = aggregatorMap.AddMax("a max key", 1)
	assert.Nil(t, err)
	err = aggregatorMap.AddMin("a min key", 90)
	assert.Nil(t, err)
	err = aggregatorMap.AddAvg("an avg key", 5)
	assert.Nil(t, err)

	// assert aggregated values
	assert.Equal(t, float64(2), aggregatorMap["a sum key"].ToNum())
	assert.Equal(t, float64(1), aggregatorMap["a max key"].ToNum())
	assert.Equal(t, float64(90), aggregatorMap["a min key"].ToNum())

	avg := aggregatorMap["an avg key"].(*Avg)
	assert.Equal(t, float64(5), avg.Sum)
	assert.Equal(t, 1, avg.Count)

	// asser type of aggregator
	assert.Equal(t, SumAggregatorType, aggregatorMap["a sum key"].Type())
	assert.Equal(t, MaxAggregatorType, aggregatorMap["a max key"].Type())
	assert.Equal(t, MinAggregatorType, aggregatorMap["a min key"].Type())
	assert.Equal(t, AvgAggregatorType, aggregatorMap["an avg key"].Type())
}

// this function tests that an error occurs if a same key is added with two different aggregators
func Test_MapAggregatorAdd_UnhappyPath(t *testing.T) {
	aggregatorMap := NewMap()
	// first add should not create an error
	err := aggregatorMap.AddSum("same key", 4)
	assert.Nil(t, err)

	// second add should create an error
	err = aggregatorMap.AddMax("same key", 7)
	assert.EqualError(t, err, "Mixed aggregators used")

	err = aggregatorMap.AddMin("same key", 7)
	assert.EqualError(t, err, "Mixed aggregators used")

	err = aggregatorMap.AddAvg("same key", 7)
	assert.EqualError(t, err, "Mixed aggregators used")
}

func Test_MapAggregatorUpdate_HappyPath(t *testing.T) {

	aggregatorMap := NewMap()
	aggregatorMap.AddSum("a sum key", 2)
	aggregatorMap.AddMax("a max key", 1)
	aggregatorMap.AddMin("a min key", 90)
	aggregatorMap.AddAvg("an avg key", 4)

	unsavedMap := NewMap()
	aggregatorMap.AddSum("a sum key", 4)
	aggregatorMap.AddMax("a max key", 65)
	aggregatorMap.AddMin("a min key", 1324)
	aggregatorMap.AddAvg("an avg key", 8)

	var wg sync.WaitGroup
	wg.Add(1)
	err := aggregatorMap.UpdateOutput(unsavedMap, &wg)
	wg.Wait()

	assert.Nil(t, err)

	// assert aggregated values
	assert.Equal(t, float64(6), aggregatorMap["a sum key"].ToNum())
	assert.Equal(t, float64(65), aggregatorMap["a max key"].ToNum())
	assert.Equal(t, float64(90), aggregatorMap["a min key"].ToNum())

	avg := aggregatorMap["an avg key"].(*Avg)
	assert.Equal(t, float64(12), avg.Sum)
	assert.Equal(t, 2, avg.Count)

	// asser type of aggregator
	assert.Equal(t, SumAggregatorType, aggregatorMap["a sum key"].Type())
	assert.Equal(t, MaxAggregatorType, aggregatorMap["a max key"].Type())
	assert.Equal(t, MinAggregatorType, aggregatorMap["a min key"].Type())
	assert.Equal(t, AvgAggregatorType, aggregatorMap["an avg key"].Type())
}
