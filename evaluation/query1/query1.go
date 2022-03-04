package query1

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/josenarvaezp/displ/pkg/aggregators"
)

const (
	L_ORDERKEY int = iota
	L_PARTKEY
	L_SUPPKEY
	L_LINENUMBER
	L_QUANTITY
	L_EXTENDEDPRICE
	L_DISCOUNT
	L_TAX
	L_RETURNFLAG
	L_LINESTATUS
	L_SHIPDATE
	L_COMMITDATE
	L_RECEIPTDATE
	L_SHIPINSTRUCT
	L_SHIPMODE
	L_COMMENT
)

func TestQuery1(filename string) aggregators.MapSum {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	output := make(aggregators.MapSum)

	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), "|")

		// shipdate_value := fields[L_SHIPDATE]

		// s_year, _ := strconv.Atoi(shipdate_value[0:4])
		// s_month, _ := strconv.Atoi(shipdate_value[6:7])
		// s_day, _ := strconv.Atoi(shipdate_value[9:10])

		// shipdate := time.Date(s_year, time.Month(s_month), s_day, 0, 0, 0, 0, time.Local)
		// beforeShipDate := time.Parse()
		// then := time.Date(1998, 12, 01, time.Local)
		// if

		// fields used to create keys
		returnflag := fields[L_RETURNFLAG]
		returnstatus := fields[L_LINESTATUS]

		// retrieve values as integers
		quantity, err := convertToInt(fields[L_QUANTITY])
		if err != nil {
			log.Fatal("error converting quantity to int")
		}

		extendedPrice, err := convertToInt(fields[L_EXTENDEDPRICE])
		if err != nil {
			log.Fatal("error converting quantity to int")
		}

		discount, err := convertToInt(fields[L_DISCOUNT])
		if err != nil {
			log.Fatal("error converting quantity to int")
		}

		tax, err := convertToInt(fields[L_TAX])
		if err != nil {
			log.Fatal("error converting quantity to int")
		}

		discPrice := extendedPrice * (1 - discount)
		charge := extendedPrice * (1 - discount) * (1 + tax)

		quantityKey := fmt.Sprintf("%s-%s-l_quantity", returnflag, returnstatus)
		basePriceKey := fmt.Sprintf("%s-%s-l_base_price", returnflag, returnstatus)
		discPriceKey := fmt.Sprintf("%s-%s-l_disc_price", returnflag, returnstatus)
		chargeKey := fmt.Sprintf("%s-%s-l_charge", returnflag, returnstatus)
		countKey := fmt.Sprintf("%s-%s-count", returnflag, returnstatus)

		// sum values
		output[quantityKey] = output[quantityKey] + quantity
		output[basePriceKey] = output[basePriceKey] + extendedPrice
		output[discPriceKey] = output[discPriceKey] + discPrice
		output[chargeKey] = output[chargeKey] + charge

		// count
		output[countKey] = output[countKey] + 1
	}

	return output
}

func convertToInt(value string) (int, error) {

	if i, err := strconv.Atoi(value); err == nil {
		return i, nil
	}

	if s, err := strconv.ParseFloat(value, 64); err == nil {
		return int(s), nil
	}

	return 0, errors.New("Error converting value")
}
