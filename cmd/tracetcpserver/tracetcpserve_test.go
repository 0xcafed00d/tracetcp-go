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
	assert.Equal(t, 1, 1)

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
