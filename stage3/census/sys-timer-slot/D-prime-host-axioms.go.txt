package load

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"path"
	"slices"
	"strings"
	"sync"
)

// The package remains the published tarball. Only the checker's served view
// corrects host declarations whose published types promise more than Node does.
//
//go:embed host_axioms.json
var hostAxiomsJSON []byte

type hostAxiom struct {
	File      string          `json:"file"`
	Line      int             `json:"line"`
	Published string          `json:"published"`
	Corrected string          `json:"corrected"`
	Node      json.RawMessage `json:"node"`
	Ruling    string          `json:"ruling"`
}

var hostAxioms = sync.OnceValues(func() ([]hostAxiom, error) {
	return parseHostAxioms(hostAxiomsJSON)
})

func parseHostAxioms(data []byte) ([]hostAxiom, error) {
	var rows []hostAxiom
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&rows); err != nil {
		return nil, fmt.Errorf("load: host_axioms.json: %w", err)
	}
	if decoder.Decode(new(any)) != io.EOF {
		return nil, fmt.Errorf("load: host_axioms.json: trailing data")
	}
	seen := map[string]bool{}
	for index, row := range rows {
		key := fmt.Sprintf("%s:%d", row.File, row.Line)
		if row.File == "." || !fs.ValidPath(row.File) || path.Ext(row.File) != ".ts" || row.Line < 1 ||
			row.Published == "" || row.Corrected == "" || row.Published == row.Corrected ||
			strings.ContainsAny(row.Published+row.Corrected, "\r\n") || len(row.Node) == 0 || row.Ruling == "" || seen[key] {
			return nil, hostAxiomError(index, row, "invalid or duplicate row")
		}
		seen[key] = true
	}
	return rows, nil
}

func hostAxiomError(index int, row hostAxiom, reason string) error {
	return fmt.Errorf("load: host axiom row %d (%s:%d): %s", index+1, row.File, row.Line, reason)
}

// NodeHostAxiomFiles names the published package files with host corrections.
// Disk-based checker preparations use the same table as the bundled view.
func NodeHostAxiomFiles() ([]string, error) {
	rows, err := hostAxioms()
	if err != nil {
		return nil, err
	}
	var files []string
	for _, row := range rows {
		if !slices.Contains(files, row.File) {
			files = append(files, row.File)
		}
	}
	return files, nil
}

// CorrectNodeHostAxioms applies line-anchored corrections to a published package
// file. It refuses drift, including an upstream fix, rather than silently serving
// an uncorrected axiom. ReadStream and WriteStream have identical isTTY lines;
// each is anchored once on its own line and explicitly accounted for in the file.
func CorrectNodeHostAxioms(file, published string) (string, error) {
	rows, err := hostAxioms()
	if err != nil {
		return "", err
	}
	return correctNodeHostAxioms(file, published, rows)
}

func correctNodeHostAxioms(file, published string, rows []hostAxiom) (string, error) {
	if !slices.ContainsFunc(rows, func(row hostAxiom) bool { return row.File == file }) {
		return published, nil
	}
	lines := strings.Split(published, "\n")
	for index, row := range rows {
		if row.File != file {
			continue
		}
		if strings.Contains(published, row.Corrected) {
			return "", hostAxiomError(index, row, "corrected text already appears in the published file; retire the row")
		}
		if row.Line > len(lines) || strings.Count(lines[row.Line-1], row.Published) != 1 {
			return "", hostAxiomError(index, row, "published text must occur exactly once on its published line")
		}
		// Count identical published lines together; an extra occurrence is unaccounted
		// drift, not a reason to replace a different declaration by spelling alone.
		expected := 0
		for _, other := range rows {
			if other.File == file && other.Published == row.Published {
				expected++
			}
		}
		if strings.Count(published, row.Published) != expected {
			return "", hostAxiomError(index, row, "published text occurrences in the file do not match the line-anchored rows")
		}
	}
	for _, row := range rows {
		if row.File == file {
			lines[row.Line-1] = strings.Replace(lines[row.Line-1], row.Published, row.Corrected, 1)
		}
	}
	return strings.Join(lines, "\n"), nil
}

func validateNodeHostAxioms(files fs.FS) error {
	rows, err := hostAxioms()
	if err != nil {
		return err
	}
	seen := map[string]bool{}
	for index, row := range rows {
		if seen[row.File] {
			continue
		}
		seen[row.File] = true
		data, err := fs.ReadFile(files, "node_types/node_modules/@types/node/"+row.File)
		if err != nil {
			return hostAxiomError(index, row, "cannot read published file: "+err.Error())
		}
		if _, err := CorrectNodeHostAxioms(row.File, string(data)); err != nil {
			return err
		}
	}
	return nil
}
