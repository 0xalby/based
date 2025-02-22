package drivers

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/charmbracelet/log"
	_ "github.com/tursodatabase/go-libsql"
)

type DriverTurso struct {
	db *sql.DB
}

func (d *DriverTurso) MustConnect(uri, user, password string) (*sql.DB, error) {
	if os.Getenv("TURSO_CLOUD") == "true" {
		// Construning the connection string
		uri = fmt.Sprintf("libsql://%s-%s.turso.io", os.Getenv("TURSO_DATABASE_NAME"), os.Getenv("TURSO_ORGANIZATION_NAME"))
	}
	var err error
	d.db, err = sql.Open("libsql", uri)
	if err != nil {
		panic(fmt.Errorf("failed to open turso database %s", err))
	}
	// Pinging to verify the connection is alive
	if err = d.db.Ping(); err != nil {
		log.Fatal("failed to ping turso database", "err", err)
		return nil, err
	}
	return d.db, nil
}

func (d *DriverTurso) Exec(directive string, args ...interface{}) (sql.Result, error) {
	result, err := d.db.Exec(directive, args...)
	if err != nil {
		return nil, fmt.Errorf("turso exec error: %v", err)
	}
	return result, nil
}

func (d *DriverTurso) Query(query string, args ...interface{}) (*sql.Rows, error) {
	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("turso query error: %v", err)
	}
	return rows, nil
}

func (d *DriverTurso) Close() error {
	if err := d.db.Close(); err != nil {
		return fmt.Errorf("turso close error: %v", err)
	}
	return nil
}
