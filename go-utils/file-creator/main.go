package main

import (
	"bytes"
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
)

// import (
// 	"encoding/csv"
// 	"flag"
// 	"fmt"
// )

const (
	SenderAddress   = "sender%d@sender%d.com"
	ReceiverAddress = "receiver%d@receiver%d.com"
	DateFormat      = "%d/%d/%d %d:%d"
	Subject         = "Hello %d"
	PolicyTypes     = "PolicyType1, PolicyType2"
	PolicyNames     = "PolicyName1, PolicyName2"
	DeliveryMethod  = "ZixPort"
)

var (
	fileName         string
	numberOfRows     int
	numberOfSpamRows int
	spamStartLine    int
	append           bool
	header           = []string{"senderAddress", "recipientAddress", "sentTimestamp", "subject", "policyTypes", "policyNames", "deliveryMethod"}
)

type Record []string

// Example of running: -output mock-zix-usage -rows 20 -spams 5 -spams-start 5 -append -
func main() {

	flag.StringVar(&fileName, "output", "", "Name of the output file")
	flag.IntVar(&numberOfRows, "rows", 0, "Number of rows")
	flag.IntVar(&numberOfSpamRows, "spams", 0, "Number of Spam")
	flag.IntVar(&spamStartLine, "spams-start", 1, "Line in which spam starts")
	flag.BoolVar(&append, "append", false, "Indicates if should append to the file or override")

	flag.Parse()

	fmt.Printf("Creating file [%s] with %d rows with %d spams (starting at line %d)\n", fileName, numberOfRows, numberOfSpamRows, spamStartLine)

	var file *os.File
	var rowsOffset = 0

	if append {
		var err error
		file, err = os.OpenFile(fmt.Sprintf("./output/%s.csv", fileName), os.O_WRONLY|os.O_APPEND, 0644)

		reader := io.Reader(file)
		rowsOffset, err = countFileLines(reader)

		fmt.Printf("Rows cont %d \n", rowsOffset)
		checkError("Cannot open file", err)
		if err != nil {
			os.Exit(1)
		}

	} else {
		var err error
		// Creating the new file
		file, err = os.Create(fmt.Sprintf("./output/%s.csv", fileName))
		checkError("Cannot create file", err)
		if err != nil {
			os.Exit(1)
		}
	}

	defer file.Close()

	csvWriter := csv.NewWriter(file)

	//Writes the header
	headerErr := csvWriter.Write(header)
	checkError("Cannot write the Header", headerErr)

	rowsCount := 0

	for i := (1 + rowsOffset); rowsCount < numberOfRows; i++ {

		sendDate := createRandomDateAsString()

		if numberOfSpamRows > 0 && spamStartLine == i {

			for y := 1; y <= numberOfSpamRows; y++ {
				writeCSVLine(i, sendDate, csvWriter)
				rowsCount++
			}

		} else {
			writeCSVLine(i, sendDate, csvWriter)
			rowsCount++
		}

	}

	// Write any buffered data to the underlying writer (standard output).
	csvWriter.Flush()

	wErr := csvWriter.Error()
	checkError("Error after flushing flushing", wErr)

}

func createRandomDateAsString() string {

	year := 2018
	month := 10
	day := rand.Intn(30) + 1

	hour := rand.Intn(12) + 1
	minute := rand.Intn(59)
	// second := rand.Intn(59)

	// var meridiem string
	// mChoose := rand.Intn(2)
	// if mChoose == 0 {
	// 	meridiem = "AM"
	// } else {
	// 	meridiem = "PM"
	// }

	return fmt.Sprintf(DateFormat, day, month, year, hour, minute)
}

func writeCSVLine(index int, sendDate string, w *csv.Writer) {
	record := Record{
		fmt.Sprintf(SenderAddress, index, index),
		fmt.Sprintf(ReceiverAddress, index, index),
		sendDate,
		fmt.Sprintf(Subject, index),
		PolicyTypes,
		PolicyNames,
		DeliveryMethod,
	}
	err := w.Write(record)
	checkError("Cannot write the record ["+record.toString()+"]", err)
}

func countFileLines(r io.Reader) (int, error) {
	buf := make([]byte, 32*1024)
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := r.Read(buf)
		fmt.Println("Count lines")
		count += bytes.Count(buf[:c], lineSep)

		switch {
		case err == io.EOF:
			return count, nil

		case err != nil:
			return count, err
		}
	}
}

func checkError(msg string, err error) {
	if err != nil {
		log.Fatalln(msg, err)
	}
}

func (r Record) toString() string {
	var buffer bytes.Buffer
	for _, column := range r {
		buffer.WriteString(column)
	}

	return buffer.String()
}
