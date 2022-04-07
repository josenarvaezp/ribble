package query6

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/josenarvaezp/displ/pkg/aggregators"
)

/*
Query 6 from https://www.tpc.org/tpc_documents_current_versions/pdf/tpc-h_v3.0.0.pdf

The Forecasting Revenue Change Query considers all the lineitems shipped in a given year with discounts between
DISCOUNT-0.01 and DISCOUNT+0.01. The query lists the amount by which the total revenue would have
increased if these discounts had been eliminated for lineitems with l_quantity less than quantity. Note that the
potential revenue increase is equal to the sum of [l_extendedprice * l_discount] for all lineitems with discounts and
quantities in the qualifying range.

select
	sum(l_extendedprice*l_discount) as revenue
from
	lineitem
where
	l_shipdate >= date '[DATE]'
	and l_shipdate < date '[DATE]' + interval '1' year
	and l_discount between [DISCOUNT] - 0.01 and [DISCOUNT] + 0.01
	and l_quantity < [QUANTITY];
*/

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

type values struct {
	shipDate      time.Time
	discount      float64
	quantity      float64
	extendedPrice float64
}

func Query6(filename string) aggregators.MapAggregator {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// init output map
	output := aggregators.NewMap()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// read line and get values
		fields := strings.Split(scanner.Text(), "|")

		if len(fields) != 17 {
			// incorrect number of fields read
			continue
		}

		lineValues := getValues(fields)

		if lineValues.skip() {
			// skip record as it doesn't accept the
			// Where statement of the query
			continue
		}

		// sum values
		output.AddSum("revenue", lineValues.extendedPrice*lineValues.discount)
	}

	return output
}

func getValues(fields []string) *values {
	shipDate := getShipDate(fields[L_SHIPDATE])

	// retrieve values as integers
	quantity, err := convertToFloat(fields[L_QUANTITY])
	if err != nil {
		log.Fatal("error converting quantity to int")
	}

	extendedPrice, err := convertToFloat(fields[L_EXTENDEDPRICE])
	if err != nil {
		log.Fatal("error converting quantity to int")
	}

	discount, err := convertToFloat(fields[L_DISCOUNT])
	if err != nil {
		log.Fatal("error converting quantity to int")
	}

	return &values{
		shipDate:      shipDate,
		quantity:      quantity,
		extendedPrice: extendedPrice,
		discount:      discount,
	}
}

func getShipDate(shipdateValue string) time.Time {
	year, _ := strconv.Atoi(shipdateValue[0:4])
	month, _ := strconv.Atoi(shipdateValue[6:7])
	day, _ := strconv.Atoi(shipdateValue[9:10])
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
}

func (v *values) skip() bool {
	dateCondition := time.Date(1994, 01, 01, 0, 0, 0, 0, time.Local)

	if v.shipDate.Before(dateCondition) {
		return true
	}

	if v.shipDate.After(dateCondition.AddDate(1, 0, 0)) || v.shipDate.Equal(dateCondition.AddDate(1, 0, 0)) {
		return true
	}

	if !(v.discount > (0.06-0.01) && v.discount < (0.06+0.01)) {
		return true
	}

	if v.quantity >= 24 {
		return true
	}

	return false
}

func convertToFloat(value string) (float64, error) {
	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, err
	}

	return floatValue, nil
}
