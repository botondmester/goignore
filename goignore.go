package goignore

import (
	"os"
	"path/filepath"
	"strings"
)

// this is my own implementation of strings.Split()
// for my use case, this is way faster than the stdlib one
// this is specialized for splitting paths, so it only splits it into
// at most 1024 components, as that should be enough for any path
func mySplit(s string, sep byte) []string {
	pathComponents := make([]string, 1024) // should be enough for all cases
	idx := 0
	for l, r := 0, 0; r <= len(s); r++ {
		if r == len(s) || s[r] == sep {
			// only add component if it is not empty
			if r > l {
				pathComponents[idx] = s[l:r]
				idx++
			}
			l = r + 1
		}
	}
	// truncate the slice to the actual number of components
	return pathComponents[:idx]
}

// Represents a single rule in a .gitignore file
// Components is a list of path components to match against
// Negate is true if the rule negates the match (i.e. starts with '!')
// OnlyDirectory is true if the rule matches only directories (i.e. ends with '/')
// Relative is true if the rule is relative (i.e. starts with '/')
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
			// just skip the character in str
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
		// TODO: handle character classes like [a-z] and [^a-z], and add tests for them
		// case '[':
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

// Tries to match the path components against the rule components
// matches is true if the path matches the rule, final is true if the rule matched the whole path
// the final parameter is used for rules that match directories only
func matchComponents(path []string, components []string, onlyDirectory bool) (matches bool, final bool) {
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
					// pass final trough
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

// Tries to match the path against the rule
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

// Stores a list of rules for matching paths against .gitignore patterns
type Gitignore struct {
	Rules []Rule
}

// Creates a Gitignore from a list of patterns (lines in a .gitignore file)
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

// Same as CompileIgnoreLines, but reads from a file
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

	// check if the pattern ends with a '/', which means it only matches directories
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

// Tries to match the path to all the rules in the gitignore
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
