package market

import (
	"encoding/csv"
	"io"
	"log"
	"os"
	"strconv"

	"github.com/josenarvaezp/displ/pkg/aggregators"
)

// MaxSales gets the maximum sales of dairy and fruits
// that where achieved in any week. Every file in
// the input bucket contains weekly sales.
func MaxSales(filename string) aggregators.MapMax {
	// create MapMax map
	output := make(aggregators.MapMax)

	// read file
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// read csv file
	csvReader := csv.NewReader(file)
	for {
		row, err := csvReader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}

		product := row[0]
		if product == "cheese" || product == "milk" {
			// convert price to int
			price, err := strconv.Atoi(row[1])
			if err != nil {
				// ignore value
			}
			output["dairy"] = output["dairy"] + price
		} else if product == "bananas" {
			// convert price to int
			price, err := strconv.Atoi(row[1])
			if err != nil {
				// ignore value
			}
			output["fruit"] = output["fruit"] + price
		}
	}

	return output
}

// MinSales gets the minimum sales of dairy and fruits
// that where achieved in any week. Every file in
// the input bucket contains weekly sales.
func MinSales(filename string) aggregators.MapMin {
	// create MapMax map
	output := make(aggregators.MapMin)

	// read file
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// read csv file
	csvReader := csv.NewReader(file)
	for {
		row, err := csvReader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}

		product := row[0]
		if product == "cheese" || product == "milk" {
			// convert price to int
			price, err := strconv.Atoi(row[1])
			if err != nil {
				// ignore value
			}
			output["dairy"] = output["dairy"] + price
		} else if product == "bananas" {
			// convert price to int
			price, err := strconv.Atoi(row[1])
			if err != nil {
				// ignore value
			}
			output["fruit"] = output["fruit"] + price
		}
	}

	return output
}
