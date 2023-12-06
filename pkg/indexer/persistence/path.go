package persistence

import (
	"path/filepath"
	"strings"
)

// SplitPath splits the path on the nodes name according to their levels:
// "/aaa" -> "aaa"
// "aaa" -> "aaa"
// "///aaa//bbb" -> "aaa", "bbb"
// "aaa//bbb/c" -> "aaa", "bbb", "c"
func SplitPath(path string) []string {
	path = strings.Trim(path, " ")
	if path == "" {
		return []string{}
	}
	res := strings.Split(filepath.Clean(path), "/")
	for len(res) > 0 && res[0] == "" {
		res = res[1:]
	}
	for len(res) > 0 && res[len(res)-1] == "" {
		res = res[:len(res)-1]
	}
	return res
}
