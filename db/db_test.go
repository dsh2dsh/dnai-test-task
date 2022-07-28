package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDB(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	c := &Config{Driver: "mysql", HostRW: "tcp(127.0.0.1)"}
	db, err := newDB("demoa", c)
	require.NoError(err)
	assert.NotNil(db.RW())
	assert.Nil(db.RO())

	c.HostRO = c.HostRW
	db, err = newDB("demoa", c)
	require.NoError(err)
	assert.NotNil(db.RO())
}

func TestAppID(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	c := &Config{Driver: "mysql", HostRW: "tcp(127.0.0.1)"}
	db, err := newDB("demoa", c)
	require.NoError(err)
	assert.Equal(db.AppID(), "demoa")
}

func TestUseInUseRelease(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	c := &Config{Driver: "mysql", HostRW: "tcp(127.0.0.1)"}
	db, err := newDB("demoa", c)
	require.NoError(err)
	assert.Equal(db.useCnt, 0)
	assert.False(db.inUse())

	db.use()
	assert.Equal(db.useCnt, 1)
	assert.True(db.inUse())

	db.release()
	assert.Equal(db.useCnt, 0)
	assert.False(db.inUse())
}

func TestClose(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	c := &Config{Driver: "mysql", HostRW: "tcp(127.0.0.1)"}
	db, err := newDB("demoa", c)
	require.NoError(err)
	assert.NotNil(db.RW())

	err = db.close()
	require.NoError(err)
	assert.Nil(db.RW())

	c.HostRO = c.HostRW
	db, err = newDB("demoa", c)
	require.NoError(err)
	assert.NotNil(db.RO())

	err = db.close()
	require.NoError(err)
	assert.Nil(db.RW())
	assert.Nil(db.RO())
}
