package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMgr(t *testing.T) {
	a := assert.New(t)
	r := require.New(t)

	withTestIdleMgr(t)
	m := NewMgr(Config{Driver: "mysql", HostRW: "tcp(127.0.0.1)"})
	r.NotNil(m)
	r.NotNil(m.idle)
	a.IsType(m.idle, newIdleMgr())
}

func TestDB(t *testing.T) {
	a := assert.New(t)
	r := require.New(t)

	withTestIdleMgr(t)
	m := NewMgr(Config{Driver: "mysql", HostRW: "tcp(127.0.0.1)"})
	r.NotNil(m)

	db, err := m.DB("demoa")
	r.NoError(err)
	r.NotNil(db)
	a.Equal(db.AppID(), "demoa")
	a.Contains(m.appDB, "demoa")

	db2 := m.appDB["demoa"]
	r.NotNil(db2)
	a.Same(db2, db)

	a.True(db.inUse())
	a.Equal(db.useCnt, 1)
}

func TestReleaseDB(t *testing.T) {
	a := assert.New(t)
	r := require.New(t)

	withTestIdleMgr(t)
	m := NewMgr(Config{Driver: "mysql", HostRW: "tcp(127.0.0.1)"})
	r.NotNil(m)

	db, err := m.DB("demoa")
	r.NoError(err)
	r.NotNil(db)

	m.ReleaseDB(db)
	a.False(db.inUse())
	a.NotContains(m.appDB, "demoa")
	a.True(m.idle.onIdle(db.AppID()))
}
