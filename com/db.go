package com

import (
	"database/sql"
	_ "embed"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed init.sql
var dbInit string

func OpenDB(dataSourceName string) (*sql.DB, error) {
	_, err := os.Stat(dataSourceName)
	init := err != nil && os.IsNotExist(err)
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, err
	}
	if init {
		if _, err = db.Exec(dbInit); err != nil {
			return nil, err
		}
	}
	return db, nil
}
