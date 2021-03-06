package main

import (
	"testing"
	"time"

	"github.com/0xcafed00d/assert"
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

	assert := assert.Make(t)

	v := Values{
		"host":     "www.google.com",
		"port":     "https",
		"starthop": "1",
		"endhop":   "30",
		"queries":  "3",
	}

	assert(parseRequest(testConfig, Values{}.read)).NoError().Equal(testConfig, nil)
	assert(parseRequest(testConfig, v.read)).NoError().Equal(testConfig, nil)

	v2 := v.clone()
	v2["starthop"] = "abc"
	assert(parseRequest(testConfig, v2.read)).HasError()

	v2 = v.clone()
	v2["endhop"] = "abc"
	assert(parseRequest(testConfig, v2.read)).HasError()

	v2 = v.clone()
	v2["queries"] = "abc"
	assert(parseRequest(testConfig, v2.read)).HasError()
}

func TestValidateConfig(t *testing.T) {

	assert := assert.Make(t)

	assert(validateConfig(&testConfig)).NoError()
	assert(validateConfig(&traceConfig{})).HasError()

	cfg := testConfig
	cfg.host += "|"
	assert(validateConfig(&cfg)).HasError()

	cfg = testConfig
	cfg.port += "&"
	assert(validateConfig(&cfg)).HasError()

	cfg = testConfig
	cfg.starthop = 0
	assert(validateConfig(&cfg)).HasError()

	cfg = testConfig
	cfg.endhop = 0
	assert(validateConfig(&cfg)).HasError()

	cfg = testConfig
	cfg.endhop = 128
	assert(validateConfig(&cfg)).HasError()

	cfg = testConfig
	cfg.endhop = 45
	cfg.starthop = 46
	assert(validateConfig(&cfg)).HasError()

	cfg = testConfig
	cfg.queries = 0
	assert(validateConfig(&cfg)).HasError()

	cfg = testConfig
	cfg.queries = 6
	assert(validateConfig(&cfg)).HasError()

	cfg = testConfig
	cfg.timeout = 3*time.Second + 1
	assert(validateConfig(&cfg)).HasError()
}

func TestCommandLine(t *testing.T) {
	assert := assert.Make(t)

	cfg := testConfig
	assert(makeCommandLine(&cfg)).Equal([]string{"-h", "1", "-m", "30", "-p", "3", "-t", "1s", "www.google.com:https"})
	cfg.nolookup = true
	assert(makeCommandLine(&cfg)).Equal([]string{"-n", "-h", "1", "-m", "30", "-p", "3", "-t", "1s", "www.google.com:https"})
}
