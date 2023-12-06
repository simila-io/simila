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

// Path composes the path from the names
func Path(names []string) string {
	if len(names) == 0 {
		return "/"
	}
	var sb strings.Builder
	for _, n := range names {
		sb.WriteString("/")
		sb.WriteString(n)
	}
	return sb.String()
}

// ConcatPath allows to concat two pathes
func ConcatPath(p1, p2 string) string {
	return filepath.Clean("/" + p1 + "/" + p2)
}
