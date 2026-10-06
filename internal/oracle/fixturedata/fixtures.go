// Package fixturedata discovers oracle entry points and their independent count records.
package fixturedata

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Fixture names a repository-relative source. Sidecars live under oracle/testdata,
// even for examples elsewhere in the repository, so there is one discovery root.
type Fixture struct {
	Path                                                  string
	Lowers, Checked, Input, Unreadable, Writes, Uncounted bool
	Arguments                                             []string
	CountsPath                                            string
	OptionsIdentity                                       string
}

type options struct {
	Path         string   `json:"path,omitempty"`
	Lowers       *bool    `json:"lowers"`
	Checked      *bool    `json:"checked"`
	Input        *bool    `json:"input"`
	Arguments    []string `json:"arguments,omitempty"`
	ArgumentsHex []string `json:"argumentsHex,omitempty"`
	Unreadable   bool     `json:"unreadable,omitempty"`
	Writes       bool     `json:"writes,omitempty"`
	Uncounted    bool     `json:"uncounted,omitempty"`
}

func decode(path string, value any) ([]byte, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(bytes.NewReader(contents))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return nil, fmt.Errorf("%s: trailing JSON", path)
	}
	return contents, nil
}

// Discover globs *.oracle.json in every directory below testdata. Only entry
// points carry sidecars: helper modules and other tests' probes are not programs
// for the differential oracle. Sorting makes execution independent of directory order.
func Discover(repository string) ([]Fixture, error) {
	root := filepath.Join(repository, "internal/oracle/testdata")
	var found []Fixture
	seen := map[string]bool{}
	err := filepath.WalkDir(root, func(directory string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() {
			return nil
		}
		paths, err := filepath.Glob(filepath.Join(directory, "*.oracle.json"))
		if err != nil {
			return err
		}
		for _, path := range paths {
			var given options
			contents, err := decode(path, &given)
			if err != nil {
				return err
			}
			if given.Lowers == nil || given.Checked == nil || given.Input == nil {
				return fmt.Errorf("%s: lowers, checked and input must be explicit booleans", path)
			}
			source := given.Path
			if source == "" {
				source, err = filepath.Rel(repository, strings.TrimSuffix(path, ".oracle.json"))
				if err != nil {
					return err
				}
				source = filepath.ToSlash(source)
			}
			if filepath.IsAbs(source) || filepath.ToSlash(filepath.Clean(source)) != source || source == ".." || strings.HasPrefix(source, "../") {
				return fmt.Errorf("%s: path must be a clean repository-relative source", path)
			}
			if extension := filepath.Ext(source); extension != ".a" && extension != ".ts" {
				return fmt.Errorf("%s: source must be .a or .ts", path)
			}
			info, err := os.Stat(filepath.Join(repository, source))
			if err != nil {
				return err
			}
			if !info.Mode().IsRegular() {
				return fmt.Errorf("%s: source is not a regular file", path)
			}
			if seen[source] {
				return fmt.Errorf("%s: duplicate fixture %s", path, source)
			}
			seen[source] = true
			if given.Arguments != nil && given.ArgumentsHex != nil {
				return fmt.Errorf("%s: use arguments or argumentsHex, not both", path)
			}
			arguments := given.Arguments
			for _, argument := range given.ArgumentsHex {
				decoded, err := hex.DecodeString(argument)
				if err != nil {
					return fmt.Errorf("%s: argumentsHex: %w", path, err)
				}
				arguments = append(arguments, string(decoded))
			}
			if !*given.Input && (len(arguments) != 0 || given.Unreadable || given.Writes) {
				return fmt.Errorf("%s: input options require input=true", path)
			}
			if *given.Input && (!*given.Lowers || *given.Checked) {
				return fmt.Errorf("%s: input fixtures must lower without an inserted check firing", path)
			}
			if *given.Checked && !*given.Lowers {
				return fmt.Errorf("%s: checked requires lowers=true", path)
			}
			found = append(found, Fixture{Path: source, Lowers: *given.Lowers, Checked: *given.Checked, Input: *given.Input,
				Arguments: arguments, Unreadable: given.Unreadable, Writes: given.Writes, Uncounted: given.Uncounted,
				CountsPath: strings.TrimSuffix(path, ".oracle.json") + ".counts.json", OptionsIdentity: string(contents)})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(found) == 0 {
		return nil, fmt.Errorf("%s: no oracle fixtures found", root)
	}
	sort.Slice(found, func(i, j int) bool { return found[i].Path < found[j].Path })
	return found, nil
}
