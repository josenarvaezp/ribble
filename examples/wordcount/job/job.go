package main

import (
	"github.com/josenarvaezp/displ/examples/wordcount"
	"github.com/josenarvaezp/displ/pkg/generators"
)

func main() {
	generators.Job(
		wordcount.WordCount,
	)
}
