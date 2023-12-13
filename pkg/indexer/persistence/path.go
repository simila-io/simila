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

// ConcatPath allows to concat two paths
func ConcatPath(p1, p2 string) string {
	return filepath.Clean("/" + p1 + "/" + p2)
}

// ToNodePath returns the path in a form as expected by Node
func ToNodePath(path string) string {
	parts := SplitPath(path)
	if len(parts) == 0 {
		return "/"
	}
	path = Path(parts)
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}
	return path
}

// ToNodePathName returns the path and name parts of the given path as expected by Node
func ToNodePathName(path string) (string, string) { // path, name
	parts := SplitPath(path)
	if len(parts) == 0 {
		return "/", ""
	}
	path = Path(parts[:len(parts)-1])
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}
	return path, strings.TrimSpace(parts[len(parts)-1])
}

// ToNodePathNamePairs returns all the possible {path, name} pairs along the given path as expected by Node
func ToNodePathNamePairs(path string) [][]string {
	pairs := make([][]string, 0)
	parts := SplitPath(path)

	var sb strings.Builder
	sb.WriteString("/")

	for i := 0; i < len(parts); i++ {
		if i > 0 {
			sb.WriteString(parts[i-1])
			sb.WriteString("/")
		}
		pairs = append(pairs, []string{sb.String(), parts[i]})
	}
	return pairs
}
