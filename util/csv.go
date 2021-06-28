package util

import (
	"encoding/csv"
	"os"
)

func ReadCSV(file string) ([][]string, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return csv.NewReader(f).ReadAll()
}

func WriteCSV(file string, records [][]string) error {
	f, err := os.OpenFile(file, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	return csv.NewWriter(f).WriteAll(records)
}
