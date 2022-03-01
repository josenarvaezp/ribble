package main

import (
	"github.com/josenarvaezp/displ/examples/wordcount"
	"github.com/josenarvaezp/displ/pkg/generators"
)

func main() {
	config := generators.Config{
		InputBuckets: []string{"input-bucket"},
		Region:       "eu-west-2",
		Local:        false,
		LogLevel:     1,
		AccountID:    "000000000000",
		Username:     "jose",
		LogicalSplit: false,
	}

	generators.Job(
		wordcount.WordCount,
		config,
	)
}
