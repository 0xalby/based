package database

import "database/sql"

type Driver interface {
	MustConnect(uri string, user string, password string) (*sql.DB, error)
	Exec(directive string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	Close() error
}
