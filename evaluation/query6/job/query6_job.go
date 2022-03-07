package main

import (
	"github.com/josenarvaezp/displ/evaluation/query6"
	"github.com/josenarvaezp/displ/pkg/ribble"
)

func main() {
	// define job's config
	config := ribble.Config{
		InputBuckets: []string{"input-bucket"},
		Region:       "eu-west-2",
		Local:        false,
		LogLevel:     1,
		AccountID:    "000000000000",
		Username:     "jose",
		LogicalSplit: true,
	}

	// define job
	ribble.Job(
		query6.TestQuery6,
		config,
	)
}
