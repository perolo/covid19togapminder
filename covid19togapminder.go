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
	"sort"
	"strconv"
	"strings"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

type csvFileType struct {
	name   string
	header []string
	lines  map[string][]string
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func main() {

	outfilePtr := flag.String("out", "tst.csv", "an output filename")
	dirPtr := flag.String("dir", ".", "a directory where all csv-files will be converted")
	parseUSPtr := flag.Bool("US", false, "set to true if electing US data, default false")

	flag.Parse()

	if len(os.Args) < 3 {
		fmt.Println("covid19togapminder -h display help")
		os.Exit(1)
	}

	ff, err := os.Create(*outfilePtr)
	check(err)
	defer ff.Close()

	var csvFiles map[string]csvFileType
	csvFiles = make(map[string]csvFileType, 10)
	files, err := ioutil.ReadDir(*dirPtr)
	check(err)
	fmt.Println("Convert Files")
	for _, fil := range files {
		if strings.HasSuffix(fil.Name(), "csv") {
			fmt.Println("  " + fil.Name())
			if strings.Contains(fil.Name(), "_US.csv") {
				if *parseUSPtr {
					fmt.Println("  US   " + fil.Name())
					var csvFile csvFileType
					csvFile.lines = make(map[string][]string)
					csvFile.name = fil.Name()
					convertUSCsvFile(*dirPtr, &csvFile)
					csvFiles[csvFile.name] = csvFile
				} else {
					fmt.Println("    Skipping  US   " + fil.Name())
				}
			} else {
				if *parseUSPtr {
					fmt.Println("    Skipping   " + fil.Name())
				} else {
					var csvFile csvFileType
					csvFile.lines = make(map[string][]string)
					csvFile.name = fil.Name()
					convertCsvFile(*dirPtr, &csvFile)
					csvFiles[csvFile.name] = csvFile
				}
			}
		}
	}
	if !*parseUSPtr {

		// Read Population data
		var popFile csvFileType
		popFile.name = "Population.dat"
		wd, err := os.Getwd()
		check(err)
		createNormData := false
		if fileExists(wd + string(os.PathSeparator) + popFile.name) {
			popFile.lines = make(map[string][]string)
			fmt.Println("Read population Data File")
			convertCsvFile(wd, &popFile)
			createNormData = true
		}

		// TODO Check same header and Data
		fmt.Println("Create Relative Data")
		relcsv := createRelCsv(csvFiles["time_series_covid19_deaths_global.csv"], csvFiles["time_series_covid19_confirmed_global.csv"], "Ratio: Death/Confirmed")
		csvFiles[relcsv.name] = relcsv
		fmt.Println("  " + relcsv.name)
		relcsv2 := createRelCsv(csvFiles["time_series_covid19_confirmed_global.csv"], csvFiles["time_series_covid19_recovered_global.csv"], "Ratio: Confirmed/Recovered")
		csvFiles[relcsv2.name] = relcsv2
		fmt.Println("  " + relcsv2.name)
		relcsv3 := createRelCsv(csvFiles["time_series_covid19_deaths_global.csv"], csvFiles["time_series_covid19_recovered_global.csv"], "Ratio: Death/Recovered")
		csvFiles[relcsv3.name] = relcsv3
		fmt.Println("  " + relcsv3.name)

		fmt.Println("Create Daily Data")
		daycsv := createDayCsv(csvFiles["time_series_covid19_deaths_global.csv"], "Day: Death")
		csvFiles[daycsv.name] = daycsv
		fmt.Println("  " + daycsv.name)
		daycsv2 := createDayCsv(csvFiles["time_series_covid19_confirmed_global.csv"], "Day: Confirmed")
		csvFiles[daycsv2.name] = daycsv2
		fmt.Println("  " + daycsv2.name)
		daycsv3 := createDayCsv(csvFiles["time_series_covid19_recovered_global.csv"], "Day: Recovered")
		csvFiles[daycsv3.name] = daycsv3
		fmt.Println("  " + daycsv3.name)

		if createNormData {
			fmt.Println("Create Population Normalized Data")
			popcsv := createNormCsv(csvFiles["time_series_covid19_deaths_global.csv"], popFile, "Population Normalized: Death")
			csvFiles[popcsv.name] = popcsv
			fmt.Println("  " + popcsv.name)
			popcsv2 := createNormCsv(csvFiles["time_series_covid19_confirmed_global.csv"], popFile, "Population Normalized: Confirmed")
			csvFiles[popcsv2.name] = popcsv2
			fmt.Println("  " + popcsv2.name)
			popcsv3 := createNormCsv(csvFiles["time_series_covid19_recovered_global.csv"], popFile, "Population Normalized: Recovered")
			csvFiles[popcsv3.name] = popcsv3
			fmt.Println("  " + popcsv3.name)
		}
	}

	fmt.Println("Write Gapminder Data")
	if *parseUSPtr {
		writeCsvFile(ff, true, csvFiles["time_series_covid19_confirmed_US.csv"])
		fmt.Println("  " + "time_series_covid19_confirmed_US.csv")
	} else {
		writeCsvFile(ff, true, csvFiles["time_series_covid19_confirmed_global.csv"])
		fmt.Println("  " + "time_series_covid19_confirmed_global.csv")
	}
	var sortednames []string
	for k := range csvFiles {
		sortednames = append(sortednames, k)
	}
	sort.Strings(sortednames)
	for _, linname := range sortednames {
		cfile := csvFiles[linname]
		if cfile.name == "time_series_covid19_confirmed_global.csv" {
			//skip
		} else if cfile.name == "time_series_covid19_confirmed_US.csv" {
			// skip
		} else {
			fmt.Println("  " + cfile.name)
			writeCsvFile(ff, false, cfile)
		}
	}
}

func createNormCsv(tcsvf csvFileType, popcsvf csvFileType, name string) csvFileType {
	var normfile csvFileType
	normfile.name = name
	normfile.lines = make(map[string][]string)
	for _, lin := range popcsvf.lines {
		dataName := lin[0]
		population := 1000000.0
		var err error
		if popcsvf.lines[dataName][2] != "" {
			population, err = strconv.ParseFloat(popcsvf.lines[dataName][2], 64)
			check(err)
		}
		if datalin, ok := tcsvf.lines[dataName]; ok {
			for i, c := range datalin {
				if i == 0 {
					normfile.lines[dataName] = []string{c}
				} else if i == 1 {
					normfile.lines[dataName] = append(normfile.lines[dataName], name)
				} else {
					var t float64
					var err error
					if c != "" {
						t, err = strconv.ParseFloat(c, 64)
						check(err)
					} else {
						t = 0
					}
					if t == 0 {
						normfile.lines[dataName] = append(normfile.lines[dataName], "0")
					} else {
						res := fmt.Sprintf("%.0f", (math.Round(1000000000.0 * t / population)))
						normfile.lines[dataName] = append(normfile.lines[dataName], res)
					}
				}
			}
		} else {
			fmt.Println("    Line Missing: " + dataName + " in " + name)
		}
	}
	return normfile
}

func createRelCsv(tcsvf csvFileType, ncsvf csvFileType, name string) csvFileType {
	var rsvfile csvFileType
	rsvfile.name = name
	rsvfile.lines = make(map[string][]string)
	for _, lin := range tcsvf.lines {
		dataName := lin[0]
		if _, ok := ncsvf.lines[dataName]; ok {
			for i, c := range lin {
				if i == 0 {
					rsvfile.lines[dataName] = []string{c}
				} else if i == 1 {
					rsvfile.lines[dataName] = append(rsvfile.lines[dataName], name)
				} else {
					var t, n float64
					var err error
					if c != "" {
						t, err = strconv.ParseFloat(c, 64)
						check(err)
					} else {
						t = 0
					}
					if ncsvf.lines[dataName][i] != "" {
						n, err = strconv.ParseFloat(ncsvf.lines[dataName][i], 64)
						check(err)
					} else {
						n = 0
					}
					if n == 0 {
						rsvfile.lines[dataName] = append(rsvfile.lines[dataName], "0")
					} else {
						res := fmt.Sprintf("%.0f", (math.Round(1000.0 * t / n)))
						rsvfile.lines[dataName] = append(rsvfile.lines[dataName], res)
					}
				}
			}
		} else {
			fmt.Println("    Line Missing: " + dataName)
		}
	}
	return rsvfile
}

func createDayCsv(tcsvf csvFileType, name string) csvFileType {
	var rsvfile csvFileType
	rsvfile.name = name
	rsvfile.lines = make(map[string][]string)
	var previous float64 = 42
	for _, lin := range tcsvf.lines {
		dataName := lin[0]
		if _, ok := tcsvf.lines[dataName]; ok {
			for i, c := range lin {
				if i == 0 {
					rsvfile.lines[dataName] = []string{c}
				} else if i == 1 {
					rsvfile.lines[dataName] = append(rsvfile.lines[dataName], name)
				} else if i == 2 {
					var err error
					if c != "" {
						previous, err = strconv.ParseFloat(c, 64)
						check(err)
					} else {
						previous = 0
					}
					rsvfile.lines[dataName] = append(rsvfile.lines[dataName], c)
				} else {
					var t float64
					var err error
					if c != "" {
						t, err = strconv.ParseFloat(c, 64)
						check(err)
					} else {
						t = 0
					}
					if t == 0 {
						rsvfile.lines[dataName] = append(rsvfile.lines[dataName], "0")
					} else {
						today := t - previous
						//						if (today < 0) {
						//							fmt.Printf("Negative Report: %s %v %.0f:\n", dataName, i, today)
						//						}
						res := fmt.Sprintf("%.0f", (math.Round(today)))
						rsvfile.lines[dataName] = append(rsvfile.lines[dataName], res)
					}
					previous = t
				}
			}
		} else {
			fmt.Println("    Line Missing: " + dataName)
		}
	}
	return rsvfile
}

func writeCsvFile(f *os.File, addheader bool, csvf csvFileType) {
	var err error
	if addheader {
		for _, h := range csvf.header {
			_, err = f.WriteString(h + ",")
			check(err)
		}
		_, err = f.WriteString("\n")
		check(err)
	}
	var sortednames []string
	for k := range csvf.lines {
		sortednames = append(sortednames, k)
	}
	sort.Strings(sortednames)
	for _, linname := range sortednames {
		lin := csvf.lines[linname]
		for _, c := range lin {
			_, err = f.WriteString(c + ",")
			check(err)
		}
		_, err = f.WriteString("\n")
		check(err)
	}
}

func convertCsvFile(adir string, csvf *csvFileType) {
	var err error
	csvfile, err := os.Open(adir + string(os.PathSeparator) + csvf.name)
	if err != nil {
		log.Fatalln("Couldn't open the csv file", err)
	}
	r := csv.NewReader(csvfile)
	line := 0
	for {
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
				csvf.lines[dataName] = append(csvf.lines[dataName], strings.TrimSuffix(csvf.name, ".csv"))
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

func convertUSCsvFile(adir string, csvf *csvFileType) {
	var err error
	var populationAvailable int = 0
	if strings.Contains(csvf.name, "time_series_covid19_deaths_US.csv") {
		populationAvailable = 1
	}
	csvfile, err := os.Open(adir + string(os.PathSeparator) + csvf.name)
	if err != nil {
		log.Fatalln("Couldn't open the csv file", err)
	}
	r := csv.NewReader(csvfile)
	line := 0
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		check(err)
		var dataName string
		for k, rec := range record {
			//assuming all files have same update date!?
			if line == 0 {
				if k < 5 {
					//nothing
					//fmt.Printf("record %s\n", rec)
				} else if k == 5 {
					csvf.header = append([]string{record[5] + "-" + record[6]})
				} else if k == 6 {
					csvf.header = append(csvf.header, "Indicator")
				} else if k >= 7 && k < (11+populationAvailable) {
					//fmt.Printf("record %s\n", rec)
					//nothing
				} else if k >= (11+populationAvailable) { //"M/D/Y"
					ti, err := dateparse.ParseLocal(rec)
					if err != nil {
						panic(err.Error())
					}
					newdate := fmt.Sprintf("%d%02d%02d", ti.Year(), ti.Month(), ti.Day())
					csvf.header = append(csvf.header, newdate)
				}
			} else if k < 5 {
				//nothing
				//fmt.Printf("record %s\n", rec)
			} else if k==5 {
				if rec == "" {
					dataName = "Other" + "-" + record[6]
				} else {
					dataName = record[5] + "-" + record[6]
				}
				csvf.lines[dataName] = []string{dataName}
			} else if k == 6 {
				csvf.lines[dataName] = append(csvf.lines[dataName], strings.TrimSuffix(csvf.name, ".csv"))
			} else if k<(11+populationAvailable) {
				//nothing
				//fmt.Printf("record %s\n", rec)
			} else {
				csvf.lines[dataName] = append(csvf.lines[dataName], record[11:]...)
				break
			}
		}
		line++
	}
}
