/*
Client contacts the server to fetch incidents details and displays in table format.
It also processes the incidents in concurrent manner and provides aggregated report.
*/

package main

import (
	"client/format/table"
	"client/incidents"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"os"
)

func main() {
	// Initialize incidents obj
	incidentsObj, err := incidents.Init()
	if err != nil {
		log.Fatal(err)
	}

	// initialize logging
	Formatter := new(log.TextFormatter)
	Formatter.TimestampFormat = "02-01-2006 15:04:05"
	Formatter.FullTimestamp = true
	log.SetFormatter(Formatter)

	// get url as first arg
	url := os.Args[1]

	// get the response using http client
	res, err := incidents.GetResponse(url)
	if err != nil {
		log.Fatal(err)
	}
	// validate response based on headers
	respErr := incidents.ValidateResponse(res)
	if respErr != nil {
		log.Fatal(respErr)
	}

	// Read the body
	err = incidentsObj.ParseBody(res)
	if err != nil {
		log.Fatal(err)
	}

	// print the table of the incidents report
	tableFmt, err := table.Format((*incidentsObj).Report)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(*tableFmt)

	// generate aggregated report based on priority
	aggReport, err := incidentsObj.GenerateAggReportPriority((*incidentsObj).Report)
	if err != nil {
		log.Fatal(err)
	}

	aggTableFmt, err := table.Format(*aggReport)
	fmt.Println(*aggTableFmt)
	if err != nil {
		log.Fatal(err)
	}

}
