package main

import (
	"fmt"

	"github.com/josenarvaezp/displ/examples/wordcount"
	gen "github.com/josenarvaezp/displ/pkg/generators"
)

func main() {
	err := gen.Generate(
		wordcount.WordCount,
	)
	if err != nil {
		fmt.Println(err)
	}
}
