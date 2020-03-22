package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"github.com/araddon/dateparse"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

type csvFileType struct {
	header []string
	lines  map[string][]string
}

func main() {

	outfilePtr := flag.String("out", "tst.csv", "an output filename")
	dirPtr := flag.String("dir", ".", "a directory where all csv-files will be converted")

	flag.Parse()

	if len(os.Args) < 3 {
		fmt.Println("covid19togapminder -h")
		os.Exit(1)
	}

	ff, err := os.Create(*outfilePtr)
	check(err)
	defer ff.Close()

	var csvFiles map[string]csvFileType
	csvFiles = make(map[string]csvFileType)
	files, err := ioutil.ReadDir(*dirPtr)
	check(err)
	for _, fil := range files {
		fmt.Println(fil.Name())
		if strings.HasSuffix(fil.Name(), "csv") {
			var csvFile csvFileType
			csvFile.lines = make(map[string][]string)
			convertCsvFile(*dirPtr, fil.Name(), &csvFile)
			csvFiles[fil.Name()] = csvFile
		}
	}
	// TODO Check same header and Data
	// Create Relative
	relcsv := createRelCsv(csvFiles["time_series_19-covid-Deaths.csv"], csvFiles["time_series_19-covid-Confirmed.csv"], "Ratio: Death/Confirmed")
	csvFiles["Ratio: Death/Confirmed"] = relcsv
	relcsv2 := createRelCsv(csvFiles["time_series_19-covid-Recovered.csv"], csvFiles["time_series_19-covid-Confirmed.csv"], "Ratio: Recovered/Confirmed")
	csvFiles["Ratio: Recovered/Confirmed"] = relcsv2

	// PrintFile
	first := true
	for _, cfile := range csvFiles {
		writeCsvFile(ff, first, &cfile)
		first = false
	}
}
func createRelCsv( tcsvf csvFileType, ncsvf csvFileType, name string) ( csvFileType){
	var rsvfile csvFileType
	rsvfile.lines = make(map[string][]string)
	for _, lin := range tcsvf.lines {
		dataName := lin[0]
		for i, c := range lin {
			if (i==0) {
				rsvfile.lines[dataName] = []string{c}
			} else if i==1 {
				rsvfile.lines[dataName] = append(rsvfile.lines[dataName], name)
			} else {
				t, err := strconv.ParseFloat(c, 64);
				check(err)
				n, err := strconv.ParseFloat(ncsvf.lines[dataName][i], 64);
				check(err)
				if (n==0) {
					rsvfile.lines[dataName] = append(rsvfile.lines[dataName], "0")
				} else {
					res := fmt.Sprintf("%.0f", (math.Round(1000.0 * t / n)))
					rsvfile.lines[dataName] = append(rsvfile.lines[dataName], res)
				}
			}
		}
	}

	return rsvfile
}

func writeCsvFile(f *os.File, addheader bool, csvf *csvFileType) {
	var err error
	if addheader {
		for _, h := range csvf.header {
			_, err = f.WriteString(h+",")
			check(err)
		}
		_, err = f.WriteString("\n")
		check(err)
	}
	for _, lin := range csvf.lines {
		for _, c := range lin {
			_, err = f.WriteString(c+",")
			check(err)
		}
		_, err = f.WriteString("\n")
		check(err)
	}
	_, err = f.WriteString("\n")
	check(err)

}

func convertCsvFile(adir string, acsvfile string, csvf *csvFileType) {
	var err error
	csvfile, err := os.Open(adir + string(os.PathSeparator) + acsvfile)
	if err != nil {
		log.Fatalln("Couldn't open the csv file", err)
	}

	r := csv.NewReader(csvfile)

	line := 0
	// Iterate through the records
	for {
		// Read each record from csv
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		check(err)
		var dataName string
		for k, rec := range record {
			//assuming all files have same update date!?
			if line == 0 {
				if k == 0 {
					csvf.header = []string{record[0] + "-" + record[1]}
				} else if k == 1 {
					csvf.header = append(csvf.header, "Indicator")
				} else if k == 2 {
					//nothing
				} else if k > 3 { //"M/D/Y"
					ti, err := dateparse.ParseLocal(rec)
					if err != nil {
						panic(err.Error())
					}
					newdate := fmt.Sprintf("%d%02d%02d", ti.Year(), ti.Month(), ti.Day())
					csvf.header = append(csvf.header, newdate)
				}
			} else if k == 0 {
				if rec == "" {
					dataName = "Country" + "-" + strings.ReplaceAll(record[1], ",", "-")
				} else {
					dataName = strings.ReplaceAll(record[0], ",", "-") + "-" + record[1]
				}
				csvf.lines[dataName] = []string{dataName}
			} else if k == 1 {
				csvf.lines[dataName] = append(csvf.lines[dataName], strings.TrimSuffix(acsvfile, ".csv"))
			} else if (k == 2) || (k == 3) {
				//nothing
			} else {
				csvf.lines[dataName] = append(csvf.lines[dataName], record[4:]...)
				break
			}
		}
		line++
	}
}
