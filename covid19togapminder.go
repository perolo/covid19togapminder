package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"github.com/araddon/dateparse"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {

	outfilePtr := flag.String("out", "tst.csv", "an output filename")
	dirPtr := flag.String("dir", ".", "a directory where all csv-files will be converted")

	flag.Parse()

	ff, err := os.Create(*outfilePtr)
	check(err)
	defer ff.Close()

	files, err := ioutil.ReadDir(*dirPtr)
	check(err)
	first := true
	for _, fil := range files {
		fmt.Println(fil.Name())
		if strings.HasSuffix(fil.Name(), "csv") {
			convertCsvFile(ff, *dirPtr,fil.Name(),first)
			first=false
		}
	}
}

func convertCsvFile(f *os.File, adir string, acsvfile string, addheader bool) {
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
		for k, rec := range record {
			//assuming all files have same update date!?
			if (addheader) {
				if (k > 3) && (line == 0) { //"M/D/Y"
					//ti, err := time.Parse("\"02/01/06\"", string(rec))
					ti, err := dateparse.ParseLocal(rec)
					if err != nil {
						panic(err.Error())
					}
					newdate := fmt.Sprintf("%d%02d%02d,", ti.Year(), ti.Month(), ti.Day())
					_, err = f.WriteString(newdate)
					check(err)
				} else if (line == 0) && (k == 0) {
					_, err = f.WriteString(record[0] + "-" + record[1] + ",")
					check(err)
				} else if (line == 0) && (k == 1) {
					_, err = f.WriteString("Indicator,")
					check(err)
				} else if (line == 0) && (k == 2) {
					//nothing
				}
			}
			if (line!=0) {
				if k == 0 {
					if rec == "" {
						_, err = f.WriteString("Country" + "-" + strings.ReplaceAll(record[1],",","-") + ",")
						check(err)
					} else {
						_, err = f.WriteString(strings.ReplaceAll(record[0],",","-")+ "-" + record[1] + ",")
						check(err)
					}
				} else if k == 1 {
					_, err = f.WriteString(strings.TrimSuffix(acsvfile,".csv" )+ ",")
					check(err)
				} else if (k == 2) || (k == 3) {
					//nothing
				} else {
					_, err = f.WriteString(rec + ",")
					check(err)
				}
			}
		}
		_, err = f.WriteString("\n")
		check(err)
		line++
	}
}
