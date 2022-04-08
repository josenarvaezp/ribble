package main

import (
	"github.com/josenarvaezp/displ/evaluation/query6"
	"github.com/josenarvaezp/displ/pkg/ribble"
)

func main() {
	// define job's config
	config := ribble.Config{
		InputBuckets:        []string{"integration-test-bucket"},
		Region:              "eu-west-2",
		Local:               true,
		LogLevel:            1,
		AccountID:           "000000000000",
		Username:            "ribble",
		LogicalSplit:        true,
		RandomizedPartition: true,
	}

	// define job
	ribble.Job(
		query6.Query6,
		nil,
		nil,
		config,
	)
}
