package goignore

import (
	"os"
	"path/filepath"
	"strings"
)

// this is my own implementation of strings.Split()
// for my use case, this is way faster than the stdlib one
// the function expects a slice of sufficient length to get passed to it,
// this avoids unnecessary memory allocation
func mySplit(s string, sep byte, pathComponentsBuf []string) []string {
	idx := 0
	sLen := len(s)
	l, r := 0, 0
	for ; r < sLen; r++ {
		if s[r] == sep {
			// only add component if it is not empty
			if r > l {
				pathComponentsBuf[idx] = s[l:r]
				idx++
			}
			l = r + 1
		}
	}

	// handle the last part separately
	if r > l {
		pathComponentsBuf[idx] = s[l:r]
		idx++
	}

	// truncate the slice to the actual number of components
	return pathComponentsBuf[:idx]
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
	// i is the index in str, j is the index in pattern
	i, j := 0, 0
	for ; i < len(str); i++ {
		if j >= len(pattern) {
			// we ran out of pattern but still have str to match
			return false
		}

		switch pattern[j] {
		case '?':
			// skip the '?' character on the pattern
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
		case '[':
			prevI := i
			prevJ := j

			j++ // skip the '[' character

			negate := false
			matched := false
			// handle special cases
			switch pattern[j] {
			case '!':
				negate = true
				j++
			case ']':
				if str[i] == ']' {
					matched = true
				}
				j++
			}

			// TODO: handle backslashes correctly
			for ; j < len(pattern) && pattern[j] != ']'; j++ {
				if matched {
					continue
				}
				if pattern[j+1] == '-' && pattern[j+2] != ']' {
					// handle ranges
					if pattern[j] <= str[i] && str[i] <= pattern[j+2] {
						matched = true
					}
				}
				if str[i] == pattern[j] {
					matched = true
				}
			}

			// revert to previous state, the '[' was just a literal
			if j == len(pattern) {
				i = prevI
				j = prevJ
				if str[i] != pattern[j] {
					return false
				}
				j++
				break
			}

			j++

			if matched == negate {
				return false
			}
		default:
			// escaping
			if pattern[j] == '\\' {
				j++
			}
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
func matchComponents(path []string, components []string) (matches bool, final bool) {
	i := 0
	for ; i < len(components); i++ {
		if i >= len(path) {
			// we ran out of path components, but still have components to match
			return false, false
		}
		if components[i] == "**" {
			// stinky recursive step
			for j := len(path) - 1; j >= i; j-- {
				match, final := matchComponents(path[j:], components[i+1:])
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
// the function expects a buffer of sufficient size to get passed to it, this avoids excessive memory allocation
func (r *Rule) matchesPath(path string, buf []string) bool {
	hasSuffix := strings.HasSuffix(path, "/")
	pathComponents := mySplit(path, '/', buf)

	if !r.Relative {
		// stinky recursive step
		for j := len(pathComponents) - 1; j >= 0; j-- {
			match, final := matchComponents(pathComponents[j:], r.Components)
			if match {
				return !r.OnlyDirectory || r.OnlyDirectory && (!final || final && hasSuffix)
			}
		}

		return false
	}

	match, final := matchComponents(pathComponents, r.Components)

	return match && (!r.OnlyDirectory || r.OnlyDirectory && (!final || final && hasSuffix))
}

// Stores a list of rules for matching paths against .gitignore patterns
// PathComponentsBuf is a temporary buffer for mySplit calls, this avoids excessive allocation
type GitIgnore struct {
	Rules             []Rule
	pathComponentsBuf []string
}

// Creates a Gitignore from a list of patterns (lines in a .gitignore file)
func CompileIgnoreLines(patterns []string) *GitIgnore {
	gitignore := &GitIgnore{
		Rules:             make([]Rule, 0, len(patterns)),
		pathComponentsBuf: make([]string, 2048),
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
func CompileIgnoreFile(filename string) (*GitIgnore, error) {
	lines, err := os.ReadFile(filename)

	if err != nil {
		return nil, err
	}
	return CompileIgnoreLines(strings.Split(string(lines), "\n")), nil
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
	// we use the default split function because this only runs once for each rule
	// this saves memory compared to using mySplit
	components := strings.Split(pattern, "/")

	return Rule{
		Components:    components,
		Negate:        negate,
		OnlyDirectory: onlyDirectory,
		Relative:      relative || len(components) > 1,
	}
}

// Tries to match the path to all the rules in the gitignore
func (g *GitIgnore) MatchesPath(path string) bool {
	path = filepath.ToSlash(path)
	matched := false

	for _, rule := range g.Rules {
		if rule.matchesPath(path, g.pathComponentsBuf) {
			if !rule.Negate {
				matched = true
			} else {
				matched = false
			}
		}
	}
	return matched
}
