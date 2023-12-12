package persistence

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSplitPath(t *testing.T) {
	assert.Equal(t, []string{}, SplitPath(""))
	assert.Equal(t, []string{}, SplitPath("   "))
	assert.Equal(t, []string{"aaa"}, SplitPath("aaa"))
	assert.Equal(t, []string{"aaa"}, SplitPath("aaa/"))
	assert.Equal(t, []string{"aaa"}, SplitPath("aaa/   "))
	assert.Equal(t, []string{}, SplitPath("/"))
	assert.Equal(t, []string{" "}, SplitPath("/ / "))
	assert.Equal(t, []string{"aaa", "b"}, SplitPath("aaa/b"))
	assert.Equal(t, []string{"aaa", "b"}, SplitPath("aaa///b"))
	assert.Equal(t, []string{"aaa", "b"}, SplitPath("//aaa///b//"))
}

func TestPath(t *testing.T) {
	assert.Equal(t, "/", Path(nil))
	assert.Equal(t, "/aaa", Path([]string{"aaa"}))
	assert.Equal(t, "/aaa/aaa", Path([]string{"aaa", "aaa"}))
}

func TestConcatPath(t *testing.T) {
	assert.Equal(t, "/", ConcatPath("", "/"))
	assert.Equal(t, "/", ConcatPath("/", ""))
	assert.Equal(t, "/", ConcatPath("", ""))
	assert.Equal(t, "/", ConcatPath("/", "/"))
	assert.Equal(t, "/aaa/bbb", ConcatPath("aaa", "bbb"))
	assert.Equal(t, "/aaa", ConcatPath("aaa", ""))
	assert.Equal(t, "/aaa", ConcatPath("", "aaa"))
	assert.Equal(t, "/aaa", ConcatPath("", "//aaa"))
}
