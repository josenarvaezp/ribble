package wordcount

import (
	"bufio"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/josenarvaezp/displ/pkg/aggregators"
)

func WordCount(filename string) aggregators.MapAggregator {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// initialize map
	output := aggregators.NewMap()

	for scanner.Scan() {
		line := scanner.Text()
		words := strings.Fields(line)
		for _, word := range words {
			output.AddSum(word, 1)
		}
	}

	return output
}

// Having filters the words that have less than 5 in their sum
func Having(mapAggregator aggregators.MapAggregator) aggregators.MapAggregator {
	// delete all items from map that have less then 5 count
	for key, aggregator := range mapAggregator {
		if aggregator.ToNum() < 5 {
			delete(mapAggregator, key)
		}
	}

	return mapAggregator
}

type AggregatorPairList []aggregators.AggregatorPair

func (p AggregatorPairList) Len() int      { return len(p) }
func (p AggregatorPairList) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p AggregatorPairList) Less(i, j int) bool {
	return p[i].Value < p[j].Value
}

// Sort sorts the output by value in ascending order
func Sort(ma aggregators.MapAggregator) sort.Interface {
	keys := make(AggregatorPairList, len(ma))
	i := 0
	for k, v := range ma {
		keys[i] = aggregators.AggregatorPair{Key: k, Value: v.ToNum()}
		i++
	}

	sort.Sort(keys)

	return keys
}
