package query1

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/josenarvaezp/displ/pkg/aggregators"
)

/*
Query 1 from https://www.tpc.org/tpc_documents_current_versions/pdf/tpc-h_v3.0.0.pdf

The Pricing Summary Report Query provides a summary pricing report for all lineitems shipped as of a given date.
The date is within 60 - 120 days of the greatest ship date contained in the database. The query lists totals for
extended price, discounted extended price, discounted extended price plus tax, average quantity, average extended
price, and average discount. These aggregates are grouped by RETURNFLAG and LINESTATUS, and listed in
ascending order of RETURNFLAG and LINESTATUS. A count of the number of lineitems in each group is
included.

select
	l_returnflag,
	l_linestatus,
	sum(l_quantity) as sum_qty,
	sum(l_extendedprice) as sum_base_price,
	sum(l_extendedprice*(1-l_discount)) as sum_disc_price,
	sum(l_extendedprice*(1-l_discount)*(1+l_tax)) as sum_charge,
	avg(l_quantity) as avg_qty,
	avg(l_extendedprice) as avg_price,
	avg(l_discount) as avg_disc,
	count(*) as count_order
from
	lineitem
where
	l_shipdate <= date '1998-12-01' - interval '[DELTA]' day (3)
group by
	l_returnflag,
	l_linestatus
order by
	l_returnflag,
	l_linestatus;
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

func Query1(filename string) aggregators.MapAggregator {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// init output map
	output := aggregators.NewMap()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), "|")

		if len(fields) != 17 {
			// incorrect number of fields read
			continue
		}

		shipdateValue := fields[L_SHIPDATE]
		year, _ := strconv.Atoi(shipdateValue[0:4])
		month, _ := strconv.Atoi(shipdateValue[6:7])
		day, _ := strconv.Atoi(shipdateValue[9:10])
		shipDate := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)

		finalDate := time.Date(1998, 12, 01, 0, 0, 0, 0, time.Local).AddDate(0, 0, -90)
		if finalDate.Before(shipDate) {
			// skip
			continue
		}

		// group by fields
		returnflag := fields[L_RETURNFLAG]
		linestatus := fields[L_LINESTATUS]

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

		tax, err := convertToFloat(fields[L_TAX])
		if err != nil {
			log.Fatal("error converting quantity to int")
		}

		discPrice := extendedPrice * (1 - discount)
		charge := extendedPrice * (1 - discount) * (1 + tax)

		sumQuantityKey := fmt.Sprintf("%s-%s-l_quantity_sum", returnflag, linestatus)
		sumBasePriceKey := fmt.Sprintf("%s-%s-l_base_price_sum", returnflag, linestatus)
		sumDiscPriceKey := fmt.Sprintf("%s-%s-l_disc_price_sum", returnflag, linestatus)
		sumChargeKey := fmt.Sprintf("%s-%s-l_charge_sum", returnflag, linestatus)
		avgQuantityKey := fmt.Sprintf("%s-%s-quantity_avg", returnflag, linestatus)
		avgPriceKey := fmt.Sprintf("%s-%s-avg_price", returnflag, linestatus)
		avgDiscKey := fmt.Sprintf("%s-%s-avg_disc", returnflag, linestatus)
		sumCountKey := fmt.Sprintf("%s-%s-count", returnflag, linestatus)

		// sum values
		output.AddSum(sumQuantityKey, quantity)
		output.AddSum(sumBasePriceKey, extendedPrice)
		output.AddSum(sumDiscPriceKey, discPrice)
		output.AddSum(sumChargeKey, charge)

		// count
		output.AddSum(sumCountKey, 1)

		// Avg values
		output.AddAvg(avgQuantityKey, quantity)
		output.AddAvg(avgPriceKey, extendedPrice)
		output.AddAvg(avgDiscKey, discount)
	}

	return output
}

func convertToFloat(value string) (float64, error) {
	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, err
	}

	// keep only two decimal places
	floatValue = (math.Round(floatValue*100) / 100)

	return floatValue, nil
}

type AggregatorPairList []aggregators.AggregatorPair

func (p AggregatorPairList) Len() int      { return len(p) }
func (p AggregatorPairList) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p AggregatorPairList) Less(i, j int) bool {
	return p[i].Key < p[j].Key
}

// Sort sorts the output by value in ascending order
func Sort(ma aggregators.MapAggregator) sort.Interface {
	keys := make(AggregatorPairList, len(ma))
	i := 0
	for k, v := range ma {
		keys[i] = aggregators.AggregatorPair{Key: k, Value: v.ToNum()}
		i++
	}

	sort.Sort(keys)

	return keys
}
