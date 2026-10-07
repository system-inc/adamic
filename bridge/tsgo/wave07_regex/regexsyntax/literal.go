package regexsyntax

import "strings"

// PatternAndFlags splits a regex literal's source text into its pattern and its flags.
//
// The text arrives as it was written, slashes included, because that is what the AST node carries.
// Splitting on the last slash rather than the first is what makes `/a\/b/g` come apart correctly:
// an escaped slash inside the pattern is still a slash to a naive scan, and the flags can only
// follow the final one.
//
// A text with no closing slash returns everything after the opening one as the pattern and no
// flags. That shape does not survive the parser, so this is defensive rather than reachable, and it
// returns the reading that keeps a caller from indexing past the end.
func PatternAndFlags(text string) (pattern string, flags string) {
	if len(text) < 2 || text[0] != '/' {
		return "", ""
	}
	lastSlash := strings.LastIndex(text[1:], "/")
	if lastSlash == -1 {
		return text[1:], ""
	}
	return text[1 : lastSlash+1], text[lastSlash+2:]
}
