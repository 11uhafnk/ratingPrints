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

type data struct {
	Name      string
	LastPrint string
	AVG       float64
	Count     int
}

func (d *data) Read(row []string) (err error) {
	if row == nil {
		panicOfError(errors.New("error read stored data"))
	}
	if len(row) != 4 {
		return errors.New("not 4 item")
	}
	d.Name = row[0]
	d.LastPrint = row[1]
	d.AVG, err = strconv.ParseFloat(row[2], 64)
	if err != nil {
		return err
	}
	d.Count, err = strconv.Atoi(row[3])
	if err != nil {
		return err
	}
	return nil
}

func (d *data) Write() (result []string) {
	result = make([]string, 4)
	result[0] = d.Name
	result[1] = d.LastPrint
	result[2] = strconv.FormatFloat(d.AVG, 'g', -1, 64)
	result[3] = strconv.Itoa(d.Count)
	return
}

// data file
// file name | date last print YYYY.MM.DD | AVG prints | count prints

func main() {

	printed := make(map[string]data)

	now := time.Now() //.Add(+time.Hour * 96 * 6) //test

	fileName := "! AAA_" + now.Format("2006") + ".csv"
	dataFile, err := os.OpenFile(fileName, os.O_RDWR, 0755)
	if !os.IsNotExist(err) {
		panicOfError(err)

		csvReader := csv.NewReader(dataFile)
		csvReader.Comma = '\t'

		header, err := csvReader.Read()
		if err != nil && err != io.EOF {
			panic(err)
		}
		if err == nil {
			if header[0] != "file name" || header[1] != "last print" || header[2] != "average print per days" || header[3] != "count prints" {
				panic(errors.New("error in header file"))
			}
		}

		for {
			row, err := csvReader.Read()
			if err == io.EOF {
				break
			}
			panicOfError(err)

			var dd data
			err = dd.Read(row)
			panicOfError(err)

			printed[dd.Name] = dd
		}

		err = dataFile.Close()
		panicOfError(err)

		err = os.Remove(fileName)
		panicOfError(err)
	}

	fl, err := os.Open(".")
	panicOfError(err)
	defer fl.Close()

	files, err := fl.Readdir(-1)
	panicOfError(err)
	for _, fi := range files {
		if fi.IsDir() ||
			fi.Name() == "! Base Counter" ||
			fi.Name() == "! Base Counter.exe" ||
			(len(fi.Name()) == 16 && fi.Name()[:8] == "printed_") ||
			(len(fi.Name()) > 0 && fi.Name()[:1] == ".") {
			fmt.Println(fi.Name(), "skip")
			continue
		}
		fmt.Println(fi)

		dd := data{fi.Name(), now.Format("2006.01.02"), 0, 1}
		if _, ok := printed[fi.Name()]; ok {
			dd = printed[fi.Name()]
		}
		fmt.Println(dd)

		tm, err := time.Parse("2006.01.02", dd.LastPrint)
		panicOfError(err)
		ii := int(now.Truncate(time.Hour*24).Sub(tm).Hours()) / 24
		fmt.Println(ii)
		if ii > 0 {
			dd.AVG = (dd.AVG*float64(dd.Count) + float64(ii)) / float64(dd.Count+1)
			dd.Count++
			dd.LastPrint = now.Format("2006.01.02")
		}
		fmt.Println(dd)
		printed[dd.Name] = dd
	}

	dataFile, err = os.Create(fileName)
	defer dataFile.Close()

	csvWriter := csv.NewWriter(dataFile)
	csvWriter.Comma = '\t'
	csvWriter.UseCRLF = true

	err = csvWriter.Write([]string{"file name", "last print", "average print per days", "count prints"})
	panicOfError(err)
	for _, dd := range printed {
		fmt.Println(dd)
		err = csvWriter.Write(dd.Write())
		panicOfError(err)
	}
	csvWriter.Flush()

	if err := csvWriter.Error(); err != nil {
		fmt.Println(err)
		panic(err)
	}
	fmt.Printf("end\n")
}
