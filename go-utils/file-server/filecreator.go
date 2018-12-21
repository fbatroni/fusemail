package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

const (
	SenderAddress       = "sender%d@sender%d.com"
	ReceiverAddress     = "receiver%d@receiver%d.com"
	DateFormat          = "%d/%d/%d %d:%d"
	Subject             = "Hello %d"
	PolicyTypes         = "PolicyType1, PolicyType2"
	PolicyNames         = "PolicyName1, PolicyName2"
	DeliveryMethod      = "ZixPort"
	dateFormatWithHours = "20060102150405"
	folder              = "output"
)

var (
	fileName      string
	numberOfRows  int
	numberDomains int
	numberSenders int
	spamStartLine int
	append        bool
	header        = []string{"senderAddress", "recipientAddress", "sentTimestamp", "subject", "policyTypes", "policyNames", "deliveryMethod"}
)

type Record []string

func CreateFile() (string, error) {

	var file *os.File
	var err error
	numberOfRows = randonNumberOfLine()
	numberDomains := randonDiffDomains()
	numberSenders := randonDiffDomains()

	if _, err := os.Stat(folder); os.IsNotExist(err) {
		os.Mkdir(folder, os.ModePerm)
	}
	err = RemoveContents(folder)
	if err != nil {
		return fileName, err
	}

	fileName = "zix-usage-data-" + time.Now().Format(dateFormatWithHours) + ".csv"

	// Creating the new file
	file, err = os.Create(filepath.Join(folder, fileName))
	if err != nil {
		return fileName, err
	}

	defer file.Close()

	csvWriter := csv.NewWriter(file)

	//Writes the header
	err = csvWriter.Write(header)
	if err != nil {
		return fileName, err
	}

	for i := 1; i < numberOfRows; i++ {

		sendDate := createRandomDateAsString()
		domainNum := randonDomains(numberDomains)
		senderNum := randonDomains(numberSenders)

		err = writeCSVLine(senderNum, domainNum, sendDate, csvWriter)
		if err != nil {
			return fileName, err
		}

	}

	// Write any buffered data to the underlying writer (standard output).
	csvWriter.Flush()

	err = csvWriter.Error()
	return fileName, err
}

func createRandomDateAsString() string {

	year := 2018
	month := 10
	day := rand.Intn(30) + 1

	hour := rand.Intn(12) + 1
	minute := rand.Intn(59)

	return fmt.Sprintf(DateFormat, day, month, year, hour, minute)
}

func randonNumberOfLine() int {
	return rand.Intn(10000000) + 1
}

func randonDiffDomains() int {
	return rand.Intn(1000) + 10
}

func randonDomains(numberOfDomains int) int {
	return rand.Intn(numberOfDomains) + 1
}

func writeCSVLine(senderNum int, domainNum int, sendDate string, w *csv.Writer) error {
	record := Record{
		fmt.Sprintf(SenderAddress, senderNum, domainNum),
		fmt.Sprintf(ReceiverAddress, domainNum, senderNum),
		sendDate,
		fmt.Sprintf(Subject, domainNum),
		PolicyTypes,
		PolicyNames,
		DeliveryMethod,
	}

	return w.Write(record)
}

func (r Record) toString() string {
	var buffer bytes.Buffer
	for _, column := range r {
		buffer.WriteString(column)
	}

	return buffer.String()
}

func RemoveContents(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}
