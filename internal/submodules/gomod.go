package submodules

import (
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
)

// A Replacement is one replace directive of a go.mod: the module it replaces, and what replaces it.
type Replacement struct {
	// ModulePath and ModuleVersion are the replaced module, left of =>. ModuleVersion is empty where the directive
	// replaces every version.
	ModulePath    string
	ModuleVersion string
	// ReplacementPath and ReplacementVersion are what replaces it, right of =>. ReplacementVersion is empty where the
	// replacement is a directory, as go requires of one, and ReplacementPath is then the directory as written,
	// relative to the go.mod's own directory unless it is absolute.
	ReplacementPath    string
	ReplacementVersion string
}

// GoModReplacements is content's replace directives, as go reads them, in file order.
//
// It reads the file as golang.org/x/mod/modfile lexes it (Go's own cmd/go vendors it): tokens split at spaces, a //
// comment to the line's end outside a string, a string quoted with " (escapes) or ` (raw) on one line, ( ) [ ] { }
// and , each a token, and => a token like any other word; a directive is one line, or a verb's ( block of lines up
// to its ). Of two directives replacing one module (its path and version), the later stands where it is and the
// earlier goes, as go mod edit's clean-up keeps them (modfile's removeDups): Adamic's go.mod replaces the shim's
// parser twice. It runs no go command: a Loom runner's stand-in go refuses go mod edit. It is the reader
// internal/buildcache's goModReplacements is, reporting every directive where that one keeps the directories alone.
func GoModReplacements(content []byte) ([]Replacement, error) {
	var replacements []Replacement
	block := ""
	for number, line := range strings.Split(string(content), "\n") {
		tokens, err := goModTokens(line)
		if err != nil {
			return nil, fmt.Errorf("line %d: %v", number+1, err)
		}
		var arguments []goModToken
		switch {
		case len(tokens) == 0:
			continue
		case block != "" && tokens[0].is(")"):
			block = ""
			continue
		case block != "":
			if block != "replace" {
				continue
			}
			arguments = tokens
		case len(tokens) == 3 && tokens[1].is("(") && tokens[2].is(")"):
			// An empty block on one line: replace ().
			continue
		case len(tokens) == 2 && tokens[1].is("("):
			block = tokens[0].text
			continue
		case tokens[0].text == "replace" && !tokens[0].quoted:
			arguments = tokens[1:]
		default:
			continue
		}
		arrow := slices.IndexFunc(arguments, func(token goModToken) bool { return token.text == "=>" && !token.quoted })
		if arrow < 1 || arrow > 2 || len(arguments) < arrow+2 || len(arguments) > arrow+3 {
			return nil, fmt.Errorf("line %d: a replace go doesn't read: %s", number+1, strings.TrimSpace(line))
		}
		replacement := Replacement{ModulePath: arguments[0].text, ReplacementPath: arguments[arrow+1].text}
		if arrow == 2 {
			replacement.ModuleVersion = arguments[1].text
		}
		if len(arguments) == arrow+3 {
			replacement.ReplacementVersion = arguments[arrow+2].text
		}
		replacements = append(replacements, replacement)
	}
	var kept []Replacement
	for index, each := range replacements {
		later := slices.ContainsFunc(replacements[index+1:], func(other Replacement) bool {
			return other.ModulePath == each.ModulePath && other.ModuleVersion == each.ModuleVersion
		})
		if !later {
			kept = append(kept, each)
		}
	}
	return kept, nil
}

// ReadGoModReplacements is GoModReplacements of the go.mod at name.
func ReadGoModReplacements(name string) ([]Replacement, error) {
	content, err := os.ReadFile(name)
	if err != nil {
		return nil, err
	}
	replacements, err := GoModReplacements(content)
	if err != nil {
		return nil, fmt.Errorf("%s: %v", name, err)
	}
	return replacements, nil
}

// ReplacementDirectory is the directory that replaces module, as its go.mod writes it. It is refused, naming what the
// go.mod does replace, when no directive replaces module; when two do, for different versions, since which one a
// build takes depends on the version it asks for; and when the replacement is another module at a version rather
// than a directory, since that names no place in the tree.
func ReplacementDirectory(replacements []Replacement, module string) (string, error) {
	var matches []Replacement
	for _, replacement := range replacements {
		if replacement.ModulePath == module {
			matches = append(matches, replacement)
		}
	}
	switch {
	case len(matches) == 0:
		var found []string
		for _, replacement := range replacements {
			found = append(found, replacement.ModulePath)
		}
		if len(found) == 0 {
			return "", fmt.Errorf("no replace directive replaces %s: the go.mod has none", module)
		}
		return "", fmt.Errorf("no replace directive replaces %s; it replaces %s", module, strings.Join(found, ", "))
	case len(matches) > 1:
		var found []string
		for _, replacement := range matches {
			found = append(found, fmt.Sprintf("%s %s => %s %s", replacement.ModulePath, replacement.ModuleVersion, replacement.ReplacementPath, replacement.ReplacementVersion))
		}
		return "", fmt.Errorf("%s is replaced %d times, by version: %s", module, len(matches), strings.Join(found, "; "))
	case matches[0].ReplacementVersion != "":
		return "", fmt.Errorf("%s is replaced by the module %s %s, not a directory", module, matches[0].ReplacementPath, matches[0].ReplacementVersion)
	}
	return matches[0].ReplacementPath, nil
}

// A goModToken is one token of a go.mod line: its text, unquoted when it was a string, and whether it was one, since a
// quoted ( or => is a path, never the punctuation or the arrow.
type goModToken struct {
	text        string
	quoted      bool
	punctuation bool
}

// is says whether the token is the punctuation mark mark.
func (token goModToken) is(mark string) bool {
	return token.punctuation && token.text == mark
}

// goModTokens is one go.mod line's tokens.
func goModTokens(line string) ([]goModToken, error) {
	var tokens []goModToken
	for index := 0; index < len(line); {
		switch c := line[index]; {
		case c == ' ' || c == '\t' || c == '\r':
			index++
		case strings.HasPrefix(line[index:], "//"):
			return tokens, nil
		case strings.ContainsRune("()[]{},", rune(c)):
			tokens = append(tokens, goModToken{text: string(c), punctuation: true})
			index++
		case c == '"' || c == '`':
			end := index + 1
			for end < len(line) && line[end] != c {
				if c == '"' && line[end] == '\\' {
					end++
				}
				end++
			}
			if end >= len(line) {
				return nil, fmt.Errorf("a string that doesn't end on its line")
			}
			unquoted, err := strconv.Unquote(line[index : end+1])
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, goModToken{text: unquoted, quoted: true})
			index = end + 1
		default:
			end := index
			for end < len(line) && !strings.ContainsRune(" \t\r()[]{},", rune(line[end])) && !strings.HasPrefix(line[end:], "//") {
				end++
			}
			tokens = append(tokens, goModToken{text: line[index:end]})
			index = end
		}
	}
	return tokens, nil
}
