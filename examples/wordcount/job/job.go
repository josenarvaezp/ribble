package main

import (
	"github.com/josenarvaezp/displ/examples/wordcount"
	"github.com/josenarvaezp/displ/pkg/ribble"
)

func main() {
	// define job's config
	config := ribble.Config{
		InputBuckets:        []string{"my-input-bucket"},
		Region:              "eu-west-2",
		Local:               true,
		LogLevel:            1,
		AccountID:           "000000000000",
		Username:            "jose",
		LogicalSplit:        true,
		RandomizedPartition: true,
	}

	// define job
	ribble.Job(
		wordcount.WordCount,
		wordcount.Having,
		wordcount.Sort,
		config,
	)
}
