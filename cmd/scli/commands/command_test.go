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

	assert.Equal(t, map[string]string{"a": "lalal", "limit": "10"}, parseParams("  a=\"lalal\"   limit=10 "))
	assert.Equal(t, map[string]string{"a": "\"lalal\"  dd", "limit": "10"}, parseParams("  a=\"lalal\"  dd   limit=10 "))
	assert.Equal(t, map[string]string{"a": "a=dd", "limit": "10"}, parseParams("  a=a\\=dd   limit=10 "))
	assert.Equal(t, map[string]string{"a=dd   limit": "10"}, parseParams("  a\\=dd   limit=10 "))
}

func TestUnqote(t *testing.T) {
	assert.Equal(t, "abcd", unquote("\"abcd\""))
	assert.Equal(t, "ab\"cd", unquote("\"ab\"cd\""))
	assert.Equal(t, "ab\"cd", unquote("    \"ab\"cd\"  "))
}

func TestSplitParams(t *testing.T) {
	assert.Equal(t, []string{""}, splitParams(""))
	assert.Equal(t, []string{""}, splitParams("    "))
	assert.Equal(t, []string{"a", "b  c", "d"}, splitParams("   a=b  c=  d"))
	assert.Equal(t, []string{"a", "b  c=  d"}, splitParams("   a=b  c\\=  d"))
}
