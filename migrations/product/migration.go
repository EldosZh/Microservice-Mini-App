package product

import (
	"database/sql"
	_ "embed"
)

//go:embed 001_init.sql
var schema string

func Apply(db *sql.DB) error { _, err := db.Exec(schema); return err }
