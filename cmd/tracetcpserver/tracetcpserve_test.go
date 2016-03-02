package main

import (
	"testing"
	"time"

	"github.com/simulatedsimian/assert"
)

type Values map[string]string

func (v Values) read(name string) (val string, ok bool) {
	val, ok = (v)[name]
	return
}

func (v Values) clone() Values {
	dst := Values{}
	for key, val := range v {
		dst[key] = val
	}
	return dst
}

var testConfig = traceConfig{
	host:     "www.google.com",
	port:     "https",
	starthop: 1,
	endhop:   30,
	timeout:  1 * time.Second,
	queries:  3,
	nolookup: false,
}

func TestParseRequest(t *testing.T) {

	v := Values{
		"host":     "www.google.com",
		"port":     "https",
		"starthop": "1",
		"endhop":   "30",
		"queries":  "3",
	}

	cfg, err := parseRequest(testConfig, Values{}.read)
	assert.Nil(t, err)
	assert.Equal(t, *cfg, testConfig)

	cfg, err = parseRequest(testConfig, v.read)
	assert.Nil(t, err)
	assert.Equal(t, *cfg, testConfig)

	v2 := v.clone()
	v2["starthop"] = "abc"
	cfg, err = parseRequest(testConfig, v2.read)
	assert.NotNil(t, err)

	v2 = v.clone()
	v2["endhop"] = "abc"
	cfg, err = parseRequest(testConfig, v2.read)
	assert.NotNil(t, err)

	v2 = v.clone()
	v2["queries"] = "abc"
	cfg, err = parseRequest(testConfig, v2.read)
	assert.NotNil(t, err)
}

func TestValidateConfig(t *testing.T) {

	assert.Nil(t, validateConfig(&testConfig))
	assert.NotNil(t, validateConfig(&traceConfig{}))

	cfg := testConfig
	cfg.host += "|"
	assert.NotNil(t, validateConfig(&cfg))

	cfg = testConfig
	cfg.port += "&"
	assert.NotNil(t, validateConfig(&cfg))

	cfg = testConfig
	cfg.starthop = 0
	assert.NotNil(t, validateConfig(&cfg))

	cfg = testConfig
	cfg.endhop = 0
	assert.NotNil(t, validateConfig(&cfg))

	cfg = testConfig
	cfg.endhop = 128
	assert.NotNil(t, validateConfig(&cfg))

	cfg = testConfig
	cfg.endhop = 45
	cfg.starthop = 46
	assert.NotNil(t, validateConfig(&cfg))

	cfg = testConfig
	cfg.queries = 0
	assert.NotNil(t, validateConfig(&cfg))

	cfg = testConfig
	cfg.queries = 6
	assert.NotNil(t, validateConfig(&cfg))

	cfg = testConfig
	cfg.timeout = 3*time.Second + 1
	assert.NotNil(t, validateConfig(&cfg))
}

func TestCommandLine(t *testing.T) {
	cfg := testConfig
	assert.Equal(t, makeCommandLine(&cfg), []string{"-h", "1", "-m", "30", "-p", "3", "-t", "1s", "www.google.com:https"})

	cfg.nolookup = true
	assert.Equal(t, makeCommandLine(&cfg), []string{"-n", "-h", "1", "-m", "30", "-p", "3", "-t", "1s", "www.google.com:https"})
}
