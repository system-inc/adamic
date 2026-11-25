package submodules

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

// A Submodule is one [submodule "<name>"] section of a .gitmodules file.
type Submodule struct {
	// Name is the section's quoted name, exactly as git reads it, escapes undone. It is the key a lookup reads by.
	Name string
	// Path is where the submodule is checked out, slash-separated. From Parse it is as its own file writes it; from
	// Read it runs from the root Read began at, with every enclosing submodule's path joined before it.
	Path string
	// URL is the url key as written. A relative url (./x or ../x) stays relative, since resolving it needs the
	// superproject's remote, which a source with no .git doesn't have.
	URL string
	// Depth is how many submodules enclose the file that declares this one: 0 for the root's own .gitmodules, 1 for
	// a submodule's, and so on.
	Depth int
	// Declared is the .gitmodules that declares this one, slash-separated from the root Read began at, so an error
	// can say which file to read. From Parse it is ".gitmodules".
	Declared string
}

// FileName is the name git gives the file in every repository and submodule.
const FileName = ".gitmodules"

// Parse reads one .gitmodules file's [submodule "<name>"] sections, in the order they appear, with no git process.
//
// It reads the part of git-config's grammar a .gitmodules file uses (git help config, "Syntax"): a section header
// [section "subsection"], whose section name is case-insensitive and whose quoted subsection is case-sensitive with
// \" and \\ escaping; a key = value line, the key case-insensitive; a value whose surrounding whitespace is dropped,
// whose "..." parts are quoted, and whose \n, \t, \b, \" and \\ escape; and comments from # or ; to the line's end
// outside quotes, on a line of their own or after a header or a value. A CRLF line and a leading byte-order mark
// read as git reads them. Keys of other sections and other keys of a submodule (branch, shallow, update) are
// skipped. A later path or url in one section stands over an earlier one, as git config --get answers.
//
// What a .gitmodules never holds is refused, never guessed: a value continued onto the next line with a trailing \,
// the deprecated [submodule.name] header, a path that is absolute or climbs out with "..", and a
// section with no path, which git itself ignores and which can't be found.
func Parse(content []byte) ([]Submodule, error) {
	content = bytes.TrimPrefix(content, []byte("\xef\xbb\xbf"))
	var submodules []Submodule
	// current is the index in submodules of the section being read, or -1 inside any other section.
	current := -1
	section := ""
	for index, line := range strings.Split(string(content), "\n") {
		number := index + 1
		line = strings.TrimSuffix(line, "\r")
		rest := strings.TrimLeft(line, " \t")
		if rest == "" || rest[0] == '#' || rest[0] == ';' {
			continue
		}
		if rest[0] == '[' {
			name, subsection, quoted, err := parseHeader(rest)
			if err != nil {
				return nil, fmt.Errorf("line %d: %v: %s", number, err, line)
			}
			section = name
			current = -1
			if section != "submodule" {
				continue
			}
			if !quoted {
				return nil, fmt.Errorf("line %d: a [submodule] header with no quoted name, which no .gitmodules this reader meets writes: %s", number, line)
			}
			submodules = append(submodules, Submodule{Name: subsection, Declared: FileName})
			current = len(submodules) - 1
			continue
		}
		key, value, hasValue, err := parseVariable(rest)
		if err != nil {
			return nil, fmt.Errorf("line %d: %v: %s", number, err, line)
		}
		if section == "" {
			return nil, fmt.Errorf("line %d: a key before any section: %s", number, line)
		}
		if current < 0 || key != "path" && key != "url" {
			continue
		}
		if !hasValue {
			return nil, fmt.Errorf("line %d: %s names no value: %s", number, key, line)
		}
		if key == "url" {
			submodules[current].URL = value
			continue
		}
		cleaned, err := cleanPath(value)
		if err != nil {
			return nil, fmt.Errorf("line %d: submodule %q: %v", number, submodules[current].Name, err)
		}
		submodules[current].Path = cleaned
	}
	for _, submodule := range submodules {
		if submodule.Path == "" {
			return nil, fmt.Errorf("submodule %q names no path", submodule.Name)
		}
	}
	return submodules, nil
}

// parseHeader reads a section header: its section name lowered, as git compares it, and its quoted subsection with
// git's escapes undone, if it has one.
func parseHeader(line string) (section, subsection string, quoted bool, err error) {
	index := 1
	for index < len(line) && (isAlphanumeric(line[index]) || line[index] == '-') {
		index++
	}
	section = strings.ToLower(line[1:index])
	if section == "" {
		return "", "", false, errors.New("a section header with no name")
	}
	if index < len(line) && line[index] == '.' {
		return "", "", false, errors.New("a [section.subsection] header, git's deprecated form, which no .gitmodules this reader meets writes")
	}
	if index < len(line) && (line[index] == ' ' || line[index] == '\t') {
		for index < len(line) && (line[index] == ' ' || line[index] == '\t') {
			index++
		}
		if index >= len(line) || line[index] != '"' {
			return "", "", false, errors.New("a section header whose subsection isn't quoted")
		}
		var name strings.Builder
		index++
		for ; index < len(line) && line[index] != '"'; index++ {
			// git drops a backslash and keeps the byte after it, whatever it is, inside a subsection.
			if line[index] == '\\' {
				index++
				if index >= len(line) {
					break
				}
			}
			name.WriteByte(line[index])
		}
		if index >= len(line) {
			return "", "", false, errors.New("a section header whose quoted name doesn't end on its line")
		}
		index++
		subsection, quoted = name.String(), true
	}
	if index >= len(line) || line[index] != ']' {
		return "", "", false, errors.New("a section header that doesn't end in ]")
	}
	if trailing := strings.TrimLeft(line[index+1:], " \t"); trailing != "" && trailing[0] != '#' && trailing[0] != ';' {
		return "", "", false, errors.New("text after a section header that isn't a comment")
	}
	return section, subsection, quoted, nil
}

// parseVariable reads a key = value line: the key lowered, as git compares it, and the value as git reads it. A key
// with no = has no value, which git reads as true.
func parseVariable(line string) (key, value string, hasValue bool, err error) {
	index := 0
	for index < len(line) && (isAlphanumeric(line[index]) || line[index] == '-') {
		index++
	}
	if index == 0 || !isLetter(line[0]) {
		return "", "", false, errors.New("a key that doesn't begin with a letter")
	}
	key = strings.ToLower(line[:index])
	rest := strings.TrimLeft(line[index:], " \t")
	if rest == "" || rest[0] == '#' || rest[0] == ';' {
		return key, "", false, nil
	}
	if rest[0] != '=' {
		return "", "", false, errors.New("a key followed by neither = nor the line's end")
	}
	value, err = parseValue(rest[1:])
	return key, value, true, err
}

// parseValue reads a value as git's config parser does: leading whitespace skipped; whitespace outside quotes kept as
// one space for each byte of it only where more of the value follows; a comment outside quotes ending it.
func parseValue(text string) (string, error) {
	var value strings.Builder
	inQuotes := false
	spaces := 0
	for index := 0; index < len(text); index++ {
		c := text[index]
		if !inQuotes && (c == ' ' || c == '\t') {
			if value.Len() > 0 {
				spaces++
			}
			continue
		}
		if !inQuotes && (c == '#' || c == ';') {
			break
		}
		value.WriteString(strings.Repeat(" ", spaces))
		spaces = 0
		switch c {
		case '"':
			inQuotes = !inQuotes
		case '\\':
			index++
			if index >= len(text) {
				return "", errors.New("a value continued onto the next line, which no .gitmodules this reader meets writes")
			}
			switch text[index] {
			case 'n':
				value.WriteByte('\n')
			case 't':
				value.WriteByte('\t')
			case 'b':
				value.WriteByte('\b')
			case '"', '\\':
				value.WriteByte(text[index])
			default:
				return "", fmt.Errorf("an escape git refuses: \\%c", text[index])
			}
		default:
			value.WriteByte(c)
		}
	}
	if inQuotes {
		return "", errors.New("a quoted value that doesn't end on its line")
	}
	return value.String(), nil
}

// cleanPath is a submodule path as a slash-separated path under the repository that declares it. One that is absolute
// or climbs out is refused, as git refuses it, since joining it to a parent's path would name a place outside the
// parent; so is one that cleans to the repository itself, which Read would walk into forever.
func cleanPath(value string) (string, error) {
	cleaned := path.Clean(value)
	switch {
	case value == "":
		return "", errors.New("an empty path")
	case path.IsAbs(cleaned) || filepath.IsAbs(value) || filepath.VolumeName(value) != "":
		return "", fmt.Errorf("the absolute path %q", value)
	case cleaned == ".":
		return "", fmt.Errorf("the path %q, which is the repository itself", value)
	case cleaned == ".." || strings.HasPrefix(cleaned, "../"):
		return "", fmt.Errorf("the path %q, which climbs out of its repository", value)
	}
	return cleaned, nil
}

func isLetter(c byte) bool {
	return 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z'
}

func isAlphanumeric(c byte) bool {
	return isLetter(c) || '0' <= c && c <= '9'
}

// Read is every submodule declared under root: root's own .gitmodules, then each submodule's own, read where the
// submodule is checked out, its paths joined to the submodule's, at every depth. root's .gitmodules must be there; a
// submodule's need not, since most submodules have none and one not checked out is an empty directory. Each file is
// read before any of the files it leads to, so the result is ordered by depth, and in file order within one file.
func Read(root string) ([]Submodule, error) {
	content, err := os.ReadFile(filepath.Join(root, FileName))
	if err != nil {
		return nil, err
	}
	top, err := Parse(content)
	if err != nil {
		return nil, fmt.Errorf("%s: %v", filepath.Join(root, FileName), err)
	}
	all := top
	for next := 0; next < len(all); next++ {
		enclosing := all[next]
		declared := path.Join(enclosing.Path, FileName)
		content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(declared)))
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, err
		}
		inner, err := Parse(content)
		if err != nil {
			return nil, fmt.Errorf("%s: %v", filepath.Join(root, filepath.FromSlash(declared)), err)
		}
		for _, submodule := range inner {
			submodule.Path = path.Join(enclosing.Path, submodule.Path)
			submodule.Depth = enclosing.Depth + 1
			submodule.Declared = declared
			all = append(all, submodule)
		}
	}
	return all, nil
}

// Find is the submodule named name among submodules, as Read returns them. The name must be declared exactly once
// across every depth: a name at no depth is refused, and so is a name declared more than once, whether one file
// declares it twice, two files at one depth each declare it, or it appears at two depths, since choosing among them
// would be a guess. Each refusal names every place it found, with its path and the .gitmodules that declares it, so
// the reader can see what moved.
func Find(submodules []Submodule, name string) (Submodule, error) {
	var matches []Submodule
	for _, submodule := range submodules {
		if submodule.Name == name {
			matches = append(matches, submodule)
		}
	}
	switch len(matches) {
	case 0:
		var found []string
		for _, submodule := range submodules {
			found = append(found, fmt.Sprintf("%q at %s (%s)", submodule.Name, submodule.Path, submodule.Declared))
		}
		sort.Strings(found)
		if len(found) == 0 {
			return Submodule{}, fmt.Errorf("no submodule is named %q: no .gitmodules read declares any submodule", name)
		}
		return Submodule{}, fmt.Errorf("no submodule is named %q; found %s", name, strings.Join(found, ", "))
	case 1:
		return matches[0], nil
	}
	var found []string
	for _, submodule := range matches {
		found = append(found, fmt.Sprintf("at %s (%s, depth %d)", submodule.Path, submodule.Declared, submodule.Depth))
	}
	return Submodule{}, fmt.Errorf("the submodule name %q appears %d times: %s", name, len(matches), strings.Join(found, ", "))
}

// Path is the path, slash-separated from root, of the submodule named name, declared by root's .gitmodules or any
// submodule's own beneath it, as Read and Find read them.
func Path(root, name string) (string, error) {
	submodules, err := Read(root)
	if err != nil {
		return "", err
	}
	submodule, err := Find(submodules, name)
	if err != nil {
		return "", fmt.Errorf("%s: %v", root, err)
	}
	return submodule.Path, nil
}
