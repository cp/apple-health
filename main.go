package main

import (
	"encoding/xml"
	"fmt"
	"github.com/araddon/dateparse"
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

func main() {
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

	record := records[0]

	fmt.Printf("last record type: %v, unit: %v, startDate: %v, value: %v", record.Type, record.Unit, record.StartDate, record.Value)
}
