package main

import (
	"database/sql"
	"encoding/xml"
	"flag"
	"fmt"
	"github.com/araddon/dateparse"
	"github.com/lib/pq"
	"io"
	"os"
	"path/filepath"
	"time"
)

type XMLRecord struct {
	XMLName    xml.Name  `xml:"Record"`
	Type       string    `xml:"type,attr"`
	SourceName string    `xml:"sourceName,attr"`
	Unit       string    `xml:"unit,attr"`
	StartDate  appleTime `xml:"startDate,attr"`
	EndDate    appleTime `xml:"endDate,attr"`
	Value      int       `xml:"value,attr"`
}

type HealthData struct {
	XMLName    xml.Name    `xml:"HealthData"`
	XMLRecords []XMLRecord `xml:"Record"`
}

type appleTime struct {
	time.Time
}

func (t *appleTime) UnmarshalXMLAttr(attr xml.Attr) error {
	parsed, err := dateparse.ParseAny(attr.Value)
	if err != nil {
		return err
	}

	*t = appleTime{parsed}
	return nil
}

func readRecords(reader io.Reader) ([]XMLRecord, error) {
	var healthData HealthData
	if err := xml.NewDecoder(reader).Decode(&healthData); err != nil {
		return nil, err
	}

	return healthData.XMLRecords, nil
}

var dbUrl = flag.String("database-url", "", "database url to connect to")

func main() {
	flag.Parse()

	filePath, err := filepath.Abs("export.xml")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	records, err := readRecords(file)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println(*dbUrl)

	db, err := sql.Open("postgres", *dbUrl)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	txn, err := db.Begin()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	stmt, err := txn.Prepare(pq.CopyIn("health_records", "date", "activity", "unit", "value", "source"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, record := range records {
		_, err = stmt.Exec(record.StartDate.Time, record.Type, record.Unit, record.Value, record.SourceName)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	_, err = stmt.Exec()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = stmt.Close()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = txn.Commit()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
