package main

import (
	"reflect"
	"testing"

	"github.com/test-go/testify/assert"
)

func TestSizeMap(t *testing.T) {
	sm := CreateSizeMap("testdata")

	keys := reflect.ValueOf(sm).MapKeys()
	assert.Equal(t, 1, len(keys))
	assert.Equal(t, 3, sm[keys[0].Int()].Len())
	assert.Equal(t, 3, sm.CountCandidates())

}
