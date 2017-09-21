package main

// Тулза для сбора статистики наличия файлов в текущей директории
// license MIT
// author 11uha 11uhafnk@gmail.com

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"
)

func panicOfError(err error) {
	if err != nil {
		panic(err)
	}
}

var (
	varHeader = []string{"File name", "Last print", "Average print per days", "Count prints", "Total count prints", "Rating"}
)

const (
	nixRunnableName = "ratingPrints"
	winRunnableName = "! RatingPrints.exe"
)

type data struct {
	Name        string
	LastPrint   string
	AVG         float64
	Prints      int
	TotalPrints int
	Rating      int
}

func main() {

	now := time.Now() //.Add(+time.Hour * 96 * 8) //test

	fileName := "! AAA_" + now.Format("2006") + ".csv"

	printed, err := readData(fileName, now)
	panicOfError(err)

	err = registrationPrinted(&printed, fileName, now)
	panicOfError(err)

	err = writeDatas(&printed, fileName)
	panicOfError(err)

	fmt.Printf("end\n")
}

// readData Чтение текущего состояния распечатанных файлов из файла "! AAA_2017.csv" где 2017 текущий год
func readData(fileName string, now time.Time) (printed map[string]data, err error) {
	printed = make(map[string]data)

	dataFile, err := os.OpenFile(fileName, os.O_RDWR, 0755)
	if !os.IsNotExist(err) {
		if err != nil {
			return printed, err
		}

		csvReader := csv.NewReader(dataFile)
		csvReader.Comma = '\t'

		header, err := csvReader.Read()
		if err != nil && err != io.EOF {
			return printed, err
		}
		if err == nil {
			if len(header) >= 4 {
				if header[0] != varHeader[0] || header[1] != varHeader[1] || header[2] != varHeader[2] || header[3] != varHeader[3] {
					return printed, errors.New("error in header file 4 columns")
				}
			}
			if len(header) >= 6 {
				if header[4] != varHeader[4] || header[5] != varHeader[5] {
					return printed, errors.New("error in header file 6 columns")
				}
			}
		}

		for {
			row, err := csvReader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				return printed, err
			}

			var dd data
			err = dd.Read(row)
			if err != nil {
				return printed, err
			}

			printed[dd.Name] = dd
		}

		err = dataFile.Close()
		if err != nil {
			return printed, err
		}

		err = os.Remove(fileName)
		if err != nil {
			return printed, err
		}
	}

	return printed, nil
}

func registrationPrinted(printed_ *map[string]data, fileName string, now time.Time) error {

	if printed_ == nil {
		return errors.New("Not printed map")
	}
	printed := *printed_

	fl, err := os.Open(".")
	if err != nil {
		return err
	}
	defer fl.Close()

	files, err := fl.Readdir(-1)
	if err != nil {
		return err
	}
	for _, fi := range files {
		if fi.IsDir() ||
			fi.Name() == nixRunnableName ||
			fi.Name() == winRunnableName ||
			(len(fi.Name()) == 14 && fi.Name()[:6] == "! AAA_") || // result files
			(len(fi.Name()) > 0 && fi.Name()[:1] == ".") {
			fmt.Println(fi.Name(), "skip")
			continue
		}
		// fmt.Println(fi)

		dd := data{fi.Name(), now.Format("2006.01.02"), 0, 1, 0, 0}
		if _, ok := printed[fi.Name()]; ok {
			dd = printed[fi.Name()]
		}
		// fmt.Println(dd)

		tm, err := time.Parse("2006.01.02", dd.LastPrint)
		if err != nil {
			return err
		}
		ii := int(now.Truncate(time.Hour*24).Sub(tm).Hours()) / 24
		// fmt.Println(ii)
		if ii > 0 {
			dd.AVG = (dd.AVG*float64(dd.Prints) + float64(ii)) / float64(dd.Prints+1)
			dd.Prints++
			dd.LastPrint = now.Format("2006.01.02")
		}
		// fmt.Println(dd)
		printed[dd.Name] = dd
	}

	*printed_ = printed
	return nil
}

func writeDatas(printed_ *map[string]data, fileName string) error {

	if printed_ == nil {
		return errors.New("Not printed map")
	}
	printed := *printed_

	dataFile, err := os.Create(fileName)
	defer dataFile.Close()

	csvWriter := csv.NewWriter(dataFile)
	csvWriter.Comma = '\t'
	csvWriter.UseCRLF = true

	err = csvWriter.Write(varHeader)
	if err != nil {
		return err
	}
	for _, dd := range printed {
		dd.TotalPrints++
		dd.Rating = dd.Prints * 100 / dd.TotalPrints
		fmt.Println(dd)
		err = csvWriter.Write(dd.Write())
		if err != nil {
			return err
		}
	}
	csvWriter.Flush()

	if err := csvWriter.Error(); err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func (d *data) Read(row []string) (err error) {
	if row == nil {
		if err != nil {
			return err
		}
	}
	if len(row) < 4 {
		return errors.New("less 4 items")
	}
	d.Name = row[0]
	d.LastPrint = row[1]
	d.AVG, err = strconv.ParseFloat(row[2], 64)
	if err != nil {
		return err
	}
	d.Prints, err = strconv.Atoi(row[3])
	if err != nil {
		return err
	}

	if len(row) >= 6 {
		d.TotalPrints, err = strconv.Atoi(row[4])
		if err != nil {
			return err
		}
		d.Rating, err = strconv.Atoi(row[5])
		if err != nil {
			return err
		}
	} else {
		d.TotalPrints = d.Prints - 1
	}

	return nil
}

func (d *data) Write() (result []string) {
	result = make([]string, 6)
	result[0] = d.Name
	result[1] = d.LastPrint
	result[2] = strconv.FormatFloat(d.AVG, 'g', -1, 64)
	result[3] = strconv.Itoa(d.Prints)
	result[4] = strconv.Itoa(d.TotalPrints)
	result[5] = strconv.Itoa(d.Rating)
	return
}
