package db

import "fmt"

// Config contains options for connecting to SQL server
type Config struct {
	Driver string // name of database driver ("mysql", "pgx", ...)
	User   string // username
	Pass   string // password
	HostRW string // [protocol[(address)]] for main (RW) connection
	// Same for optional replica connection. Shoul be empty string if not used.
	HostRO string
}

// hasRO returns do Config has defined HostRO
func (self *Config) hasRO() bool {
	return self.HostRO != ""
}

// host returns content of HostRW (true) or HostRO (false) depending on value of
// rw.
func (self *Config) host(rw bool) string {
	if rw {
		return self.HostRW
	} else {
		return self.HostRO
	}
}

// FormatDSN formats the given Config into a DSN string which can be passed to
// the driver.
//
// dbName is database name we are connecting to.
//
// rw defines are we connecting to HostRW (true) or HostRO (false).
func (self *Config) formatDSN(dbName string, rw bool) string {
	return fmt.Sprintf("%s:%s@%s/%s", self.User, self.Pass, self.host(rw), dbName)
}
