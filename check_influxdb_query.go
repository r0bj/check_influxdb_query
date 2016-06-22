package main

import (
	"fmt"
	"encoding/json"

	"github.com/influxdata/influxdb/client/v2"
	"github.com/olorin/nagiosplugin"
	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	ver string = "0.10"
)

var (
	host = kingpin.Flag("host", "influxdb host").Default("localhost").Short('H').String()
	port = kingpin.Flag("port", "influxdb port").Default("8086").Short('P').String()
	username = kingpin.Flag("username", "influxdb username").Default("admin").Short('u').String()
	password = kingpin.Flag("password", "influxdb password").Default("admin").Short('p').String()
	db = kingpin.Flag("db", "influxdb database name").Default("telegraf").Short('d').String()
	warningThreshold = kingpin.Flag("warning-threshold", "warning threshold for returned value").Short('w').Required().Int()
	criticalThreshold = kingpin.Flag("critical-threshold", "critical threshold for returned value").Short('c').Required().Int()
	compareOperator = kingpin.Flag("compare-operator", "operator to compare returned value with thresholds, 'lt' or 'gt'").Short('o').Default("lt").String()
	query = kingpin.Arg("query", "influxdb query which returns one value to be able compare against integer thresholds").Required().String()
)

func queryDB(c client.Client, db string, cmd string) (res []client.Result, err error) {
	q := client.Query{
		Command:  cmd,
		Database: db,
	}
	if response, err := c.Query(q); err == nil {
		if response.Error() != nil {
			return res, response.Error()
		}
		res = response.Results
	} else {
		return res, err
	}
	return res, nil
}

func main() {
	kingpin.Version(ver)
	kingpin.Parse()

	check := nagiosplugin.NewCheck()
	defer check.Finish()

	if *compareOperator == "gt" {
		if *warningThreshold < *criticalThreshold {
			check.AddResult(nagiosplugin.UNKNOWN, "warning threshold lower than critical")
			return
		}
	} else if *compareOperator == "lt" {
		if *warningThreshold > *criticalThreshold {
			check.AddResult(nagiosplugin.UNKNOWN, "warning threshold higher than critical")
			return
		}
	}
	if *warningThreshold == *criticalThreshold {
		check.AddResult(nagiosplugin.UNKNOWN, "warning and critical thresholds are equal")
		return
	}

	if *compareOperator != "lt" && *compareOperator != "gt" {
		check.AddResult(nagiosplugin.UNKNOWN, "compare-operator parameter should be 'lt' or 'gt'")
		return	
	}

	url := fmt.Sprintf("http://%s:%s", *host, *port)
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr: url,
		Username: *username,
		Password: *password,
	})
	if err != nil {
		check.AddResult(nagiosplugin.UNKNOWN, "influxdb connect error")
		return
	}

	res, err := queryDB(c, *db, *query)
	if err != nil {
		check.AddResult(nagiosplugin.UNKNOWN, "influxdb query failed")
		return
	}

	if len(res[0].Series) < 1 {
		check.AddResult(nagiosplugin.UNKNOWN, "influxdb query returns no value")
		return	
	}

	if len(res[0].Series) > 1 {
		check.AddResult(nagiosplugin.UNKNOWN, "influxdb query returns more that one series")
		return	
	}

	if len(res[0].Series[0].Values) > 1 {
		check.AddResult(nagiosplugin.UNKNOWN, "influxdb query returns more that one value")
		return	
	}

	count := res[0].Series[0].Values[0][1]

	var i float64
	err = json.Unmarshal([]byte(count.(json.Number)), &i)
	if err != nil {
		check.AddResult(nagiosplugin.UNKNOWN, "unmarshal error")
		return
	}

	var warn float64 = float64(*warningThreshold)
	var crit float64 = float64(*criticalThreshold)

	if *compareOperator == "gt" {
		if i > warn {
			check.AddResult(nagiosplugin.OK, fmt.Sprintf("value %.2f is above thresholds", i))
		} else if i < warn && i > crit {
			check.AddResult(nagiosplugin.WARNING, fmt.Sprintf("value %.2f is below warning threshold", i))
		} else if i < crit {
			check.AddResult(nagiosplugin.CRITICAL, fmt.Sprintf("value %.2f is below critical threshold", i))
		}
	} else if *compareOperator == "lt" {
		if i < warn {
			check.AddResult(nagiosplugin.OK, fmt.Sprintf("value %.2f is below thresholds", i))
		} else if i > warn && i < crit {
			check.AddResult(nagiosplugin.WARNING, fmt.Sprintf("value %.2f is above warning threshold", i))
		} else if i > crit {
			check.AddResult(nagiosplugin.CRITICAL, fmt.Sprintf("value %.2f is above critical threshold", i))
		}
	}
}
