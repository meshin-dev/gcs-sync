package ignore

import (
	"path/filepath"
	"regexp"
	"strings"
)

// globToRegex converts a glob pattern to a regular expression string.
//
// The function handles the following glob syntax:
//   - '**': Matches any number of directories
//   - '*': Matches any number of characters except '/'
//   - '?': Matches any single character
//
// Parameters:
//   - glob: A string representing the glob pattern to be converted.
//
// Returns:
//
//	A string representing the equivalent regular expression pattern,
//	with '^' at the start and '$' at the end to ensure full string matching.
func globToRegex(glob string) string {
	re := regexp.QuoteMeta(glob)
	re = strings.ReplaceAll(re, `\*\*`, `.*`)
	re = strings.ReplaceAll(re, `\*`, `[^/]*`)
	re = strings.ReplaceAll(re, `\?`, `.`)
	return "^" + re + "$"
}

// Compile converts a slice of glob patterns into a slice of compiled regular expressions.
// It ensures that all patterns use unix-style path separators and are cleaned before compilation.
//
// Parameters:
//   - root: A string representing the root directory. This parameter is currently unused
//     but may be intended for future use in path manipulation.
//   - patterns: A slice of strings, each representing a glob pattern to be compiled.
//
// Returns:
//   - A slice of *regexp.Regexp, each corresponding to a compiled pattern.
//   - An error if any pattern fails to compile into a valid regular expression.
func Compile(root string, patterns []string) ([]*regexp.Regexp, error) {
	var res []*regexp.Regexp
	for _, g := range patterns {
		// ensure unix-style path separators inside regex
		p := filepath.ToSlash(filepath.Clean(g))
		re, err := regexp.Compile(globToRegex(p))
		if err != nil {
			return nil, err
		}
		res = append(res, re)
	}
	return res, nil
}

// Match checks if a given relative path matches any of the provided regular expressions.
//
// This function iterates through the slice of regular expressions and returns true
// if the relative path matches any of them. It's typically used to determine if a
// file or directory should be ignored based on a set of patterns.
//
// Parameters:
//   - rel: A string representing the relative path to check. This should be in unix-style
//     format and relative to the rule root.
//   - regexes: A slice of compiled regular expressions (*regexp.Regexp) to match against.
//
// Returns:
//
//	A boolean value. True if the relative path matches any of the regular expressions,
//	false otherwise.
func Match(rel string, regexes []*regexp.Regexp) bool {
	for _, re := range regexes {
		if re.MatchString(rel) {
			return true
		}
	}
	return false
}
