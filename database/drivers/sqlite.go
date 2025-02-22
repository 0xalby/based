package drivers

import (
	"database/sql"
	"fmt"

	"github.com/charmbracelet/log"
	"modernc.org/sqlite"
)

type DriverSqlite3 struct {
	db *sql.DB
}

func (d *DriverSqlite3) MustConnect(uri, user, password string) (*sql.DB, error) {
	var err error
	d.db, err = sql.Open("sqlite", uri)
	if err != nil {
		panic(fmt.Errorf("failed to open sqlite database %s", err))
	}
	// Pinging to verify the connection is alive
	if err = d.db.Ping(); err != nil {
		log.Fatal("failed to ping sqlite database", "err", err)
		return nil, err
	}
	return d.db, nil
}

func (d *DriverSqlite3) Exec(directive string, args ...interface{}) (sql.Result, error) {
	result, err := d.db.Exec(directive, args...)
	if err != nil {
		if sqliteErr, ok := err.(*sqlite.Error); ok {
			return nil, fmt.Errorf("sqlite error %d %v", sqliteErr.Code(), sqliteErr.Error())
		} else {
			return nil, err
		}
	}
	return result, nil
}

func (d *DriverSqlite3) Query(query string, args ...interface{}) (*sql.Rows, error) {
	rows, err := d.db.Query(query, args...)
	if err != nil {
		if sqliteErr, ok := err.(*sqlite.Error); ok {
			return nil, fmt.Errorf("sqlite error %d %v", sqliteErr.Code(), sqliteErr.Error())
		} else {
			return nil, err
		}
	}
	return rows, nil
}

func (d *DriverSqlite3) Close() error {
	if err := d.db.Close(); err != nil {
		return err
	}
	return nil
}
