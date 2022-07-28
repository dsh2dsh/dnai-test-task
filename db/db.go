package db

import (
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

// newDB creates and returns pool of DB connections to database with name in
// appID. It creates and initialize an instance of [DB]. In case of errors it
// returns error, else - nil as an error.
//
// dbConfig contains data for connecting to SQL server.
func newDB(appID string, dbConfig *Config) (*DB, error) {
	dbRW, err := sqlx.Open(dbConfig.Driver, dbConfig.formatDSN(appID, true))
	if err != nil {
		return nil, err
	}
	db := &DB{appID: appID, dbRW: dbRW}

	if dbConfig.hasRO() {
		dbRO, err := sqlx.Open(dbConfig.Driver, dbConfig.formatDSN(appID, false))
		if err != nil {
			dbRW.Close()
			return nil, err
		}
		db.dbRO = dbRO
		configureDB(db.RO())
	}
	configureDB(db.RW())

	return db, nil
}

// configureDB configures pool of DB connections db before using it
func configureDB(db *sqlx.DB) {
	db.SetConnMaxLifetime(3 * time.Minute)
}

// DB defines pool of connections to a database of specific appID. It's safe to
// call methods from different goroutines.
type DB struct {
	// Name of database and ID of this app
	appID string
	// Pool of read-write connections to the DB
	dbRW *sqlx.DB
	// Pool of read-only connections. May be nil if we don't have replicas.
	dbRO *sqlx.DB
	mu   sync.RWMutex
	// How many goroutines use this object at this moment
	useCnt int
}

// AppID returns application ID for which this DB was created
func (self *DB) AppID() string {
	return self.appID
}

// use just thread-safe increases num of active usage of this DB. It marks this
// DB has one more active usage. Should be called every time before usage.
func (self *DB) use() {
	self.mu.Lock()
	self.useCnt++
	self.mu.Unlock()
}

// inUse returns true if this DB is still active. It means someone uses it at
// this moment.
func (self *DB) inUse() bool {
	self.mu.RLock()
	cnt := self.useCnt
	self.mu.RUnlock()
	return cnt > 0
}

// release marks we don't use this DB anymore. Should be called every time we
// don't need this DB any more. When nobody else uses it, inUse will return
// false.
func (self *DB) release() {
	self.mu.Lock()
	self.useCnt--
	self.mu.Unlock()
}

// RW returns [*sqlx.DB] pool of connections to read-write server. Should be
// always non nil.
func (self *DB) RW() *sqlx.DB {
	return self.dbRW
}

// RO returns [*sqlx.DB] pool of connections to read-only replica, if
// any. Because replica is optional, RO can return nil in this case.
func (self *DB) RO() *sqlx.DB {
	return self.dbRO
}

// close closes all [*sqlx.DB] pools. Returns error or nil. It isn't safe to
// call it concurrently.
func (self *DB) close() error {
	if self.RO() != nil {
		if err := self.RO().Close(); err != nil {
			return err
		}
		self.dbRO = nil
	}

	if err := self.RW().Close(); err != nil {
		return err
	}
	self.dbRW = nil

	return nil
}
