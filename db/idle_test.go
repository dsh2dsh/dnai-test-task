package db

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestIdleMgrRunner struct {
	// Closing of closeCh means we've done
	closeCh chan struct{}
	// number of calls of the run method
	cnt int
}

// fake run method
func (self *TestIdleMgrRunner) run() {
	self.cnt++
	close(self.closeCh)
}

func withTestIdleMgr(t *testing.T) {
	orig := goIdleMgr
	t.Cleanup(func() { goIdleMgr = orig })
	goIdleMgr = func(m idleMgrRunner) {}
}

func TestNewIdleMgr(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	orig := goIdleMgr
	withTestIdleMgr(t)
	testRunner := &TestIdleMgrRunner{
		closeCh: make(chan struct{}),
	}
	goIdleMgr = func(m idleMgrRunner) { orig(testRunner) }
	m := newIdleMgr()
	require.NotNil(m)

	select {
	case <-testRunner.closeCh:
	case <-time.After(time.Second):
		assert.FailNow("run() method wasn't called")
	}
	assert.Equal(testRunner.cnt, 1)

	assert.Equal(m.expInterval, defExpirationInterval)
	assert.Equal(m.expJitter, defExpirationJitter)
	assert.Equal(m.maxTTL, defMaxTTL)
}

func TestOnIdle(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	withTestIdleMgr(t)
	m := newIdleMgr()
	require.NotNil(m)

	c := &Config{Driver: "mysql", HostRW: "tcp(127.0.0.1)"}
	db, err := newDB("demoa", c)
	require.NoError(err)
	assert.False(m.onIdle(db.AppID()))

	m.idleAppDB(db)
	assert.True(m.onIdle(db.AppID()))
	assert.Same(m.AppDB(db.AppID()), db)
	assert.False(m.onIdle(db.AppID()))
}

func TestSleepTime(t *testing.T) {
	assert := assert.New(t)

	expInt, _ := time.ParseDuration("45s")
	m := &idleMgr{
		expInterval: expInt,
		expJitter:   30,
	}
	rand.Seed(time.Now().UnixNano())
	sleepTime := m.sleepTime()
	assert.GreaterOrEqual(sleepTime, m.expInterval)

	max, _ := time.ParseDuration("1m15s")
	assert.Less(sleepTime, max)
}

func TestExpire(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	withTestIdleMgr(t)
	m := newIdleMgr()
	require.NotNil(m)

	c := &Config{Driver: "mysql", HostRW: "tcp(127.0.0.1)"}
	db, err := newDB("demoa", c)
	require.NoError(err)

	m.maxTTL = 0
	m.idleAppDB(db)
	assert.True(m.onIdle(db.AppID()))
	m.expire()
	assert.False(m.onIdle(db.AppID()))

	m.maxTTL = time.Minute
	m.idleAppDB(db)
	assert.True(m.onIdle(db.AppID()))
	m.expire()
	assert.True(m.onIdle(db.AppID()))
}
