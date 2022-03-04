package wordcount

import (
	"bufio"
	"log"
	"os"
	"strings"

	"github.com/josenarvaezp/displ/pkg/aggregators"
)

func WordCount(filename string) aggregators.MapSum {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	output := make(aggregators.MapSum)
	for scanner.Scan() {
		line := scanner.Text()
		words := strings.Fields(line)
		for _, word := range words {
			output[word] = output[word] + 1
		}
	}

	return output
}

func SingleWordCount(filename string) aggregators.Sum {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	output := aggregators.Sum(0)
	for scanner.Scan() {
		line := scanner.Text()
		words := strings.Fields(line)
		for _, word := range words {
			if word == "hello" {
				output = output + aggregators.Sum(1)
			}
		}
	}

	return output
}
