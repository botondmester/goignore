package goignore

import (
	"os"
	"path/filepath"
	"strings"
)

// this is my own implementation of strings.Split()
// for my use case, this is way faster than the stdlib one
func mySplit(s string, sep byte) []string {
	pathComponents := make([]string, 1024) // should be enough for all cases
	idx := 0
	for l, r := 0, 0; r <= len(s); r++ {
		if r == len(s) || s[r] == sep {
			if r > l {
				pathComponents[idx] = s[l:r]
				idx++
			}
			l = r + 1
		}
	}
	return pathComponents[:idx]
}

type Rule struct {
	Components    []string
	Negate        bool
	OnlyDirectory bool
	Relative      bool
}

func stringMatch(str string, pattern string) bool {
	i, j := 0, 0
	for ; i < len(str); i++ {
		if j >= len(pattern) {
			return false
		}

		switch pattern[j] {
		case '?':
			j++
		case '*':
			// stinky recursive step
			found := false
			for k := len(str); k >= i; k-- {
				if stringMatch(str[k:], pattern[j+1:]) {
					found = true
					break
				}
			}
			return found
		default:
			if str[i] != pattern[j] {
				return false
			}
			j++
		}
	}
	if j < len(pattern)-1 {
		// we ran out of str, but still have pattern to match
		return false
	}
	return true
}

func matchComponents(path []string, components []string, onlyDirectory bool) (bool, bool) {
	i := 0
	for ; i < len(components); i++ {
		if i >= len(path) {
			// we ran out of path components, but still have components to match
			return false, false
		}
		if components[i] == "**" {
			// stinky recursive step
			for j := len(path) - 1; j >= i; j-- {
				match, final := matchComponents(path[j:], components[i+1:], onlyDirectory)
				if match {
					return true, final
				}
			}
			return false, false
		}

		if !stringMatch(path[i], components[i]) {
			return false, false
		}
	}
	return true, i == len(path) // if we matched all components, check if we are at the end of the path
}

func (r *Rule) Match(path string) bool {
	hasSuffix := strings.HasSuffix(path, "/")
	pathComponents := mySplit(path, '/')

	if !r.Relative {
		// stinky recursive step
		for j := len(pathComponents) - 1; j >= 0; j-- {
			match, final := matchComponents(pathComponents[j:], r.Components, r.OnlyDirectory)
			if match {
				return !r.OnlyDirectory || r.OnlyDirectory && (!final || final && hasSuffix)
			}
		}

		return false
	}

	match, final := matchComponents(pathComponents, r.Components, r.OnlyDirectory)

	return match && (!r.OnlyDirectory || r.OnlyDirectory && (!final || final && hasSuffix))
}

type Gitignore struct {
	Rules []Rule
}

func CompileIgnoreLines(patterns []string) *Gitignore {
	gitignore := &Gitignore{
		Rules: make([]Rule, 0, len(patterns)),
	}

	for _, pattern := range patterns {
		// skip empty lines, comments, and trailing/leading whitespace
		pattern = strings.Trim(pattern, " \t\r\n")
		if pattern == "" || pattern[0] == '#' {
			continue
		}

		rule := createRule(pattern)

		gitignore.Rules = append(gitignore.Rules, rule)
	}

	return gitignore
}

func CompileIgnoreFile(filename string) (*Gitignore, error) {
	lines, err := os.ReadFile(filename)

	return CompileIgnoreLines(strings.Split(string(lines), "\n")), err
}

// create a rule from a pattern
func createRule(pattern string) Rule {
	negate := false
	onlyDirectory := false
	relative := false
	if pattern[0] == '!' {
		negate = true
		pattern = pattern[1:] // skip the '!'
	}

	if pattern[0] == '/' {
		relative = true
		pattern = pattern[1:] // skip the '/'
	}

	if pattern[0] == '\\' {
		pattern = pattern[1:] // skip the '\'
	}

	if pattern[len(pattern)-1] == '/' {
		onlyDirectory = true
	}

	// split the pattern into components
	components := mySplit(pattern, '/')

	return Rule{
		Components:    components,
		Negate:        negate,
		OnlyDirectory: onlyDirectory,
		Relative:      relative || len(components) > 1,
	}
}

func (g *Gitignore) Match(path string) bool {
	path = filepath.ToSlash(path)
	matched := false
	for _, rule := range g.Rules {
		if rule.Match(path) {
			if !rule.Negate {
				matched = true
			} else {
				matched = false
			}
		}
	}
	return matched
}
