// Package gatesample selects the pinned file sample used by landing gates.
// An unset switch leaves local and continuous full-gate corpora unchanged.
package gatesample

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// Selection contains repository-relative paths. Samples are sorted; full mode
// preserves the caller's order.
type Selection struct {
	Paths                 []string
	Sample                bool
	Total, Stride, Offset int
}

// Log returns the landing gate's stable diagnostic. Call it once per sampled test.
func (s Selection) Log(testName string) string {
	return fmt.Sprintf("gate-sample: %s %d/%d files, stride %d, offset %d", testName, len(s.Paths), s.Total, s.Stride, s.Offset)
}

// Select reads the landing switches and selects from physical corpus files.
// Use Generated to add an indexed generated corpus; retain mutants separately.
// controls names physical witnesses that must be checked at every offset.
// root locates packages for changed-path fixture inclusion; paths are relative to it.
func Select(root string, paths []string, stride int, controls ...string) (Selection, error) {
	sha, enabled := os.LookupEnv("ADAMIC_GATE_SAMPLE")
	selection := Selection{Paths: append([]string(nil), paths...), Total: len(paths), Stride: stride, Sample: enabled}
	if !enabled {
		return selection, nil
	}
	if err := Validate(); err != nil {
		return Selection{}, err
	}
	if stride < 1 {
		return Selection{}, fmt.Errorf("gate sample stride must be positive: %d", stride)
	}
	prefix, _ := strconv.ParseUint(sha[:8], 16, 32)
	selection.Offset = int(prefix % uint64(stride))
	changed := map[string]bool{}
	for _, name := range controls {
		changed[name] = true
	}
	packages := map[string]bool{}
	if list := os.Getenv("ADAMIC_GATE_CHANGED"); list != "" {
		data, err := os.ReadFile(list)
		if err != nil {
			return Selection{}, fmt.Errorf("ADAMIC_GATE_CHANGED: %w", err)
		}
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSuffix(line, "\r")
			if line == "" {
				continue
			}
			if filepath.IsAbs(line) || strings.HasPrefix(filepath.ToSlash(filepath.Clean(line)), "../") || filepath.Clean(line) == ".." {
				return Selection{}, fmt.Errorf("ADAMIC_GATE_CHANGED: not a repository path: %q", line)
			}
			name := filepath.ToSlash(filepath.Clean(line))
			changed[name] = true
			// Fixture Go adapters belong to the package above testdata/fixtures,
			// not to an apparent package inside that ignored corpus tree.
			directory := filepath.Dir(name)
			parts := strings.Split(name, "/")
			for index, part := range parts {
				if part == "testdata" || part == "fixtures" {
					directory = filepath.Join(parts[:index]...)
					if directory == "" {
						directory = "."
					}
					break
				}
			}
			for {
				entries, err := os.ReadDir(filepath.Join(root, directory))
				if err != nil && !os.IsNotExist(err) {
					return Selection{}, fmt.Errorf("changed package %s: %w", directory, err)
				}
				found := false
				for _, entry := range entries {
					if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".go") {
						found = true
						break
					}
				}
				if found {
					packages[filepath.ToSlash(directory)] = true
					break
				}
				if directory == "." {
					break
				}
				directory = filepath.Dir(directory)
			}
		}
	}
	sorted := append([]string(nil), paths...)
	sort.Strings(sorted)
	selection.Paths = nil
	for index, name := range sorted {
		include := index%stride == selection.Offset || changed[name]
		for directory := range packages {
			prefix := directory + "/"
			if directory == "." {
				prefix = ""
			}
			if strings.HasPrefix(name, prefix+"testdata/") || strings.HasPrefix(name, prefix+"fixtures/") {
				include = true
			}
		}
		if include {
			selection.Paths = append(selection.Paths, name)
		}
	}
	return selection, nil
}

// Validate rejects a malformed switch even when an optional oracle is absent.
func Validate() error {
	sha, enabled := os.LookupEnv("ADAMIC_GATE_SAMPLE")
	if !enabled {
		return nil
	}
	if len(sha) != 40 {
		return fmt.Errorf("ADAMIC_GATE_SAMPLE=%q: expected 40 hex characters", sha)
	}
	if _, err := hex.DecodeString(sha); err != nil {
		return fmt.Errorf("ADAMIC_GATE_SAMPLE=%q: expected 40 hex characters", sha)
	}
	return nil
}

// Generated adds a deterministic generated corpus keyed by its enumeration index.
// controls are indices of explicit checks which run at every offset. Physical
// changed-path inclusion has already been resolved by Select.
func (s Selection) Generated(count int, controls ...int) Selection {
	keep := make(map[int]bool, len(controls))
	for _, index := range controls {
		keep[index] = true
	}
	s.Total += count
	for index := 0; index < count; index++ {
		if !s.Sample || index%s.Stride == s.Offset || keep[index] {
			s.Paths = append(s.Paths, GeneratedKey(index))
		}
	}
	return s
}

// GeneratedKey identifies an input without depending on its contents or label.
func GeneratedKey(index int) string { return fmt.Sprintf("generated-index/%d", index) }
