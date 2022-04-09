package main

import (
	"github.com/josenarvaezp/displ/build/integration_tests/tests/query1"
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
		RandomizedPartition: false,
	}

	// define job
	ribble.Job(
		query1.Query1,
		nil,
		query1.Sort,
		config,
	)
}
