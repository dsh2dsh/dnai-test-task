// Package db implements global manager of DB connections per app.
//
// It's safe to use it from goroutines.
//
// For any app we can request DB connections and manager will return it. When we
// finished processing of request for this app, we should return its DB
// connections back to the manager and if nobody else uses it at this moment,
// the manager will put it into idle list and it'll be closed after some time,
// if nobody will request it before.
package db

import (
	"sync"

	"golang.org/x/sync/singleflight"
)

// NewMgr creates the manager and returns it. After that it's ready to use and
// fully functional. Should be done once at the beginning.
func NewMgr(dbConfig Config) *Mgr {
	return &Mgr{
		dbConfig: &dbConfig,
		appDB:    make(map[string]*DB),
		idle:     newIdleMgr(),
	}
}

// Mgr defines the manager. Use [NewMgr] for creating instance of Mgr.
type Mgr struct {
	dbConfig *Config // DB connection configuration
	// Creates link between ID of app and pool of DB connections to its database
	appDB map[string]*DB
	idle  *idleMgr
	mu    sync.RWMutex
	sg    singleflight.Group
}

// DB returns [DB] pools for this appID. Also it returns error if any or
// nil. When we don't need this DB anymore, it should be returned back to the
// manager by [ReleaseDB].
func (self *Mgr) DB(appID string) (*DB, error) {
	self.mu.RLock()
	if db, ok := self.appDB[appID]; ok {
		db.use()
		self.mu.RUnlock()
		return db, nil
	}
	self.mu.RUnlock()

	_, err, _ := self.sg.Do(appID, func() (any, error) {
		return self.maybeIdleDB(appID)
	})
	if err != nil {
		return nil, err
	}

	return self.DB(appID)
}

// maybeIdleDB returns [DB] from [idleMgr], and error if any or nil, for
// specified appID. Returns nil instead of [DB] if this appID isn't registered
// in [idleMgr].
func (self *Mgr) maybeIdleDB(appID string) (*DB, error) {
	db := self.idle.AppDB(appID)
	if db == nil {
		dbn, err := newDB(appID, self.dbConfig)
		if err != nil {
			return nil, err
		}
		db = dbn
	}

	self.mu.Lock()
	self.appDB[appID] = db
	self.mu.Unlock()

	return db, nil
}

// ReleaseDB returns db back into the manager. Should be called every time we
// don't need db anymore, at the end of processing. If nobody else uses this db
// at this moment, it'll be put into an idle list and later will be closed, if
// nobody else will request it before.
func (self *Mgr) ReleaseDB(db *DB) {
	db.release()
	self.mu.Lock()
	if !db.inUse() {
		self.idle.idleAppDB(db)
		delete(self.appDB, db.AppID())
	}
	self.mu.Unlock()
}
