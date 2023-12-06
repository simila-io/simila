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
