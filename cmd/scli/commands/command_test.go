package commands

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseParams(t *testing.T) {
	assert.Equal(t, map[string]string{}, parseParams(""))
	assert.Equal(t, map[string]string{}, parseParams("   "))
	assert.Equal(t, map[string]string{"a": "b"}, parseParams("  a=b "))
	assert.Equal(t, map[string]string{"a": "1234", "c": "d"}, parseParams("  a=1234   c=d "))
	assert.Equal(t, map[string]string{"a": "{aaa: bbb}", "limit": "10"}, parseParams("  a={aaa: bbb}   limit=10 "))
	assert.Equal(t, map[string]string{"a": "{aaa: bbb}  limit"}, parseParams("  a={aaa: bbb}  limit "))
}
