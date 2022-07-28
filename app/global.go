// Package app defines global state for our application.
//
// We create it at the beginning and use it all the time. It holds all things
// needed for processing every request.
package app

import (
	"os"

	"dsh/px/db"
)

// New creates instance of [Global] and returns it. Expects next env variables:
//
//   * DB_DRIVER:  name of database driver ("mysql", "pgx", ...)
//   * DB_USER:    connection username
//   * DB_PASS:    connection password
//   * DB_HOST_RW: [protocol[(address)]] for main (RW) connection
//   * DB_HOST_RO:
//
//     Same for optional replica connection. Shoul be empty string if not used.
func New() *Global {
	dbConfig := db.Config{
		Driver: os.Getenv("DB_DRIVER"),
		User:   os.Getenv("DB_USER"),
		Pass:   os.Getenv("DB_PASS"),
		HostRW: os.Getenv("DB_HOST_RW"),
		HostRO: os.Getenv("DB_HOST_RO"),
	}
	return &Global{
		db: db.NewMgr(dbConfig),
	}
}

// Global is our global state
type Global struct {
	// Manager of DB pools
	db *db.Mgr
}
