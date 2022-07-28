package db

import (
	"container/list"
	"log"
	"math/rand"
	"sync"
	"time"
)

const (
	// Default interval between expiration loops in [time.Duration]. How often we
	// are executing the expiration loop.
	defExpirationInterval = 45 * time.Second

	// Default jitter for [defExpirationInterval] in seconds. We add random num of
	// seconds to the interval before each sleep in the expiration loop.
	defExpirationJitter = 30

	// Default Time-To-Life for idle entries in [time.Duration]. Every idle [DB]
	// will be closed after max TTL in this state.
	defMaxTTL = 5 * time.Minute
)

// idleMgrRunner is the interface what wraps the run method.
//
// run launches forever loop, which cleanups idle DBs and sleeps between
// cleanups.
type idleMgrRunner interface {
	run()
}

// goIdleMgr launches a goroutine with the run method of idleMgr. We are
// modifying it in tests.
var goIdleMgr = func(m idleMgrRunner) {
	go m.run()
}

// newIdleMgr creates, initializes and returns manager, which keeps and handles
// idle [DB]. This manager is thread-safe.
func newIdleMgr() *idleMgr {
	m := &idleMgr{
		idleList:    list.New(),
		idleMap:     make(map[string]*list.Element),
		expInterval: defExpirationInterval,
		expJitter:   defExpirationJitter,
		maxTTL:      defMaxTTL,
	}
	goIdleMgr(m)
	return m
}

// idleMgr is manager of idle [DB]. Its methods can be called from different
// goroutines. It keeps [DB] nobody uses for [maxTTL] time and closes it if
// nobody asked before.
type idleMgr struct {
	// List of idle [DB] ordered from fresh to old. If last element is not enough
	// old according to [maxTTL], it means every item before is not reached
	// [maxTTL] too.
	idleList *list.List

	// Just a map between appID and its [DB] for faster access
	idleMap map[string]*list.Element

	// Interval between expiration loops
	expInterval time.Duration

	// Jitter, which we use to add random numbers of seconds to expiration
	// interval on every iteration of the expiration loop.
	expJitter int

	// Max Time-To-Life of idle [DB]. If nobody will get it, we'll close it.
	maxTTL time.Duration

	mu sync.RWMutex
}

// idleDB is a Value of [list.Element]. We are saving here pointer to [DB] and
// expiration time.
type idleDB struct {
	db       *DB
	expireAt time.Time
}

// onIdle returns true if idleMgr contains [DB] for appID. It means appID's DB
// is idle, nobody uses it and it'll be closed soon.
func (self *idleMgr) onIdle(appID string) bool {
	self.mu.RLock()
	_, ok := self.idleMap[appID]
	self.mu.RUnlock()
	return ok
}

// AppDB returns DB for appID if idleMgr contains it or nil. It removes DB from
// idle state and forgets about it.
func (self *idleMgr) AppDB(appID string) *DB {
	self.mu.RLock()

	if _, ok := self.idleMap[appID]; ok {
		self.mu.RUnlock()
		self.mu.Lock()
		defer self.mu.Unlock()
		if elem, ok := self.idleMap[appID]; ok {
			idle := elem.Value.(idleDB)
			delete(self.idleMap, appID)
			self.idleList.Remove(elem)
			return idle.db
		}
	} else {
		self.mu.RUnlock()
	}

	return nil
}

// idleAppDB adds db into idleMgr and marks it for expiration after maxTTL. It
// adds db into front of the list and it guaranties us the list is always sorted
// from freshest to oldest.
func (self *idleMgr) idleAppDB(db *DB) {
	self.mu.Lock()
	defer self.mu.Unlock()

	appID := db.AppID()
	ttl := time.Now().UTC().Add(self.maxTTL)
	if elem, ok := self.idleMap[appID]; ok {
		idle := elem.Value.(idleDB)
		idle.expireAt = ttl
	} else {
		elem = self.idleList.PushFront(idleDB{db, ttl})
		self.idleMap[appID] = elem
	}
}

// run starts the expiration loop. On every iteration it sleeps for expInterval
// + random number of seconds, defined by expJitter. Works in separate goroutine
// and fired from newIdleMgr.
func (self *idleMgr) run() {
	for {
		time.Sleep(self.sleepTime())
		self.expire()
	}
}

// sleepTime returns current sleep time before calling the expire method. Every
// call returns expInterval + some random num of seconds < expJitter.
func (self *idleMgr) sleepTime() time.Duration {
	jitter := time.Duration(rand.Intn(self.expJitter) * int(time.Second))
	return self.expInterval + jitter
}

// expire closes and removes all DB with expireAt before now. Because our list
// by nature sorted from freshest to oldest, we can stop our loop as soon as the
// last element is fresh enough.
func (self *idleMgr) expire() {
	self.mu.Lock()
	defer self.mu.Unlock()

	now := time.Now().UTC()
	for elem := self.idleList.Back(); elem != nil; elem = self.idleList.Back() {
		idle := elem.Value.(idleDB)
		if idle.expireAt.After(now) {
			return
		}
		db := idle.db
		delete(self.idleMap, db.AppID())
		self.idleList.Remove(elem)
		if err := db.close(); err != nil {
			log.Printf("expire: close idle DB pool(%v): %v\n", db.AppID(), err)
		}
	}
}
