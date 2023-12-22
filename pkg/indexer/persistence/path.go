package persistence

import (
	"fmt"
	"github.com/acquirecloud/golibs/errors"
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
	ln := 0
	for _, n := range names {
		ln += len(n)
	}
	ln += len(names)
	var sb strings.Builder
	sb.Grow(ln)
	for _, n := range names {
		sb.WriteString("/")
		sb.WriteString(n)
	}
	return sb.String()
}

// CleanName checks whether the name contains '/' and trims spaces if needed. It returns the modified name or
// an error if any
func CleanName(name string) (string, error) {
	pieces := SplitPath(name)
	if len(pieces) != 1 {
		return name, fmt.Errorf("the name %q is incorrect, it cannot contain / symbols: %w", name, errors.ErrInvalid)
	}
	if pieces[0] == "" {
		return name, fmt.Errorf("the name %q is an empty string, not allowed: %w", name, errors.ErrInvalid)
	}
	return pieces[0], nil
}

// ConcatPath allows to concat two paths
func ConcatPath(p1, p2 string) string {
	var sb strings.Builder
	sb.Grow(len(p1) + len(p2) + 2)
	sb.WriteString("/")
	sb.WriteString(p1)
	sb.WriteString("/")
	sb.WriteString(p2)
	return filepath.Clean(sb.String())
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
