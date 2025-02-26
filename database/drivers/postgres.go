package drivers

import (
	"database/sql"
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/lib/pq"
)

type DriverPostgres struct {
	db *sql.DB
}

func (d *DriverPostgres) MustConnect(uri, user, password string) (*sql.DB, error) {
	// Constructing the connection string
	connStr := fmt.Sprintf("postgres://%s:%s@%s", user, password, uri)
	// Opening the connection
	var err error
	d.db, err = sql.Open("postgres", connStr)
	if err != nil {
		panic(fmt.Errorf("failed to open postgres database %s", err))
	}
	// Pinging to verify the connection is alive
	if err = d.db.Ping(); err != nil {
		log.Fatal("failed to ping postgres database", "err", err)
		return nil, err
	}
	return d.db, nil
}

func (d *DriverPostgres) Exec(directive string, args ...interface{}) (sql.Result, error) {
	result, err := d.db.Exec(directive, args...)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok {
			return nil, fmt.Errorf("postgres error %s %v", pgErr.Code, pgErr.Error())
		} else {
			return nil, err
		}
	}
	return result, nil
}

func (d *DriverPostgres) Query(query string, args ...interface{}) (*sql.Rows, error) {
	rows, err := d.db.Query(query, args...)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok {
			return nil, pgErr
		} else {
			return nil, err
		}
	}
	return rows, nil
}

func (d *DriverPostgres) Close() error {
	if err := d.db.Close(); err != nil {
		if pgErr, ok := err.(*pq.Error); ok {
			return pgErr
		}
		return err
	}
	return nil
}
