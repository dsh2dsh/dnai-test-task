package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHasRO(t *testing.T) {
	assert := assert.New(t)

	c := &Config{}
	assert.False(c.hasRO())
	c.HostRO = "something"
	assert.True(c.hasRO())
}

func TestHost(t *testing.T) {
	assert := assert.New(t)

	c := &Config{HostRW: "tcp(db1)", HostRO: "tcp(db2)"}
	assert.Equal(c.host(true), c.HostRW, "Didn't get value of HostRW")
	assert.Equal(c.host(false), c.HostRO, "Didn't get value of HostRO")
}

func TestFormatDSN(t *testing.T) {
	assert := assert.New(t)

	c := &Config{
		User:   "user",
		Pass:   "password",
		HostRW: "tcp(db1)",
		HostRO: "tcp(db2)",
	}
	assert.Equal(c.formatDSN("test", true), "user:password@tcp(db1)/test")
	assert.Equal(c.formatDSN("test", false), "user:password@tcp(db2)/test")
}
