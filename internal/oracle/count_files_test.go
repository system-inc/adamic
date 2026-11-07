package oracle

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// Variant records name an explicit source; unknown labels still fail the ordinary file census.
var countFixtureSources = map[string]string{}

// Keep the full source extension: foo.a and foo.ts are distinct fixtures.
func countFile(kind, key string) string {
	return filepath.Join(countsDirectory, kind, filepath.FromSlash(key)+".counts")
}

func countRecord(kind, key, row string) ([]byte, error) {
	if !filepath.IsLocal(filepath.FromSlash(key)) ||
		filepath.ToSlash(filepath.Clean(filepath.FromSlash(key))) != key ||
		strings.Contains(key, "\\") {
		return nil, fmt.Errorf("invalid count fixture key %q", key)
	}
	columns := 7
	if kind == "predicates" {
		columns = 5
	} else if kind != "runtime" {
		return nil, fmt.Errorf("unknown count kind %q", kind)
	}
	fields := strings.Split(row, " | ")
	if len(fields) != columns || fields[0] != "| "+key || !strings.HasSuffix(fields[len(fields)-1], " |") {
		return nil, fmt.Errorf("invalid %s count row for %s: %q", kind, key, row)
	}
	fields[len(fields)-1] = strings.TrimSuffix(fields[len(fields)-1], " |")
	for _, field := range fields[1:] {
		value, parseError := strconv.ParseUint(field, 10, 64)
		if parseError != nil || strconv.FormatUint(value, 10) != field {
			return nil, fmt.Errorf("invalid count %q for %s", field, key)
		}
	}
	return []byte(row + "\n"), nil
}

func readCountRow(root, kind, key string) (string, error) {
	file := filepath.Join(root, kind, filepath.FromSlash(key)+".counts")
	data, readError := os.ReadFile(file)
	if readError != nil {
		return "", readError
	}
	row := strings.TrimSuffix(string(data), "\n")
	canonical, countRecordError := countRecord(kind, key, row)
	if countRecordError != nil {
		return "", fmt.Errorf("%s: %w", file, countRecordError)
	}
	if !bytes.Equal(data, canonical) {
		return "", fmt.Errorf("%s must contain exactly one row with an LF newline", file)
	}
	return row, nil
}

func readCountRows(root, kind string, keys []string) (map[string]string, error) {
	rows := map[string]string{}
	for _, key := range keys {
		if _, exists := rows[key]; exists {
			return nil, fmt.Errorf("duplicate %s count fixture %s", kind, key)
		}
		row, readCountRowError := readCountRow(root, kind, key)
		if readCountRowError != nil {
			return nil, readCountRowError
		}
		rows[key] = row
	}
	return rows, nil
}

// The updater and temporary regeneration use the same host-independent encoding.
// The caller supplies actual measurements, not newly inferred expected counts.
func writeCountFiles(root string, runtime, predicates map[string]string) error {
	files := map[string][]byte{}
	for kind, rows := range map[string]map[string]string{"runtime": runtime, "predicates": predicates} {
		for key, row := range rows {
			data, countRecordError := countRecord(kind, key, row)
			if countRecordError != nil {
				return countRecordError
			}
			files[filepath.Join(root, kind, filepath.FromSlash(key)+".counts")] = data
		}
	}
	paths := make([]string, 0, len(files))
	for path := range files {
		paths = append(paths, path)
	}
	slices.Sort(paths)
	for _, path := range paths {
		data := files[path]
		if old, readError := os.ReadFile(path); readError == nil && bytes.Equal(old, data) {
			continue
		}
		if createDirectoryError := os.MkdirAll(filepath.Dir(path), 0o755); createDirectoryError != nil {
			return createDirectoryError
		}
		if writeError := os.WriteFile(path, data, 0o644); writeError != nil {
			return writeError
		}
	}
	return nil
}

func countsFileCensus(root string) error {
	predicateKeys, predicateCountKeysError := predicateCountKeys()
	if predicateCountKeysError != nil {
		return predicateCountKeysError
	}
	expected := map[string]bool{}
	for kind, keys := range map[string][]string{"runtime": countKeys(), "predicates": predicateKeys} {
		for _, key := range keys {
			path := filepath.Join(root, kind, filepath.FromSlash(key)+".counts")
			if expected[path] {
				return fmt.Errorf("duplicate counts file %s", path)
			}
			expected[path] = true
			fixture := key
			if source, ok := countFixtureSources[key]; ok {
				fixture = source
			}
			if info, statError := os.Stat(
				filepath.Join(repository, filepath.FromSlash(fixture)),
			); statError != nil || !info.Mode().IsRegular() {
				return fmt.Errorf("%s has no fixture %s: %v", path, key, statError)
			}
			if _, readCountRowError := readCountRow(root, kind, key); readCountRowError != nil {
				return fmt.Errorf("fixture %s needs counts file %s: %w", key, path, readCountRowError)
			}
		}
	}
	return filepath.WalkDir(root, func(path string, entry fs.DirEntry, pathError error) error {
		if pathError != nil {
			return pathError
		}
		if entry.IsDir() || path == filepath.Join(root, "README.md") {
			return nil
		}
		if !expected[path] {
			return fmt.Errorf("orphan counts file %s has no counted fixture", path)
		}
		return nil
	})
}

func TestCountsFilesCensus(test *testing.T) {
	test.Parallel()
	// New files are published by TestMain after a successful measurement run.
	if *updateCounts {
		return
	}
	if countsFileCensusError := countsFileCensus(countsDirectory); countsFileCensusError != nil {
		test.Fatal(countsFileCensusError)
	}
}

// Each existing count shard regenerates independently measured values through the
// updater's writer, in temporary storage, and compares committed bytes. No fixture
// is measured by a monolithic extra test that could exceed its four-CPU budget.
func regenerateCountRecord(test *testing.T, kind, key, measured string) {
	test.Helper()
	root := test.TempDir()
	rows := map[string]string{key: measured}
	var writeCountFilesError error
	if kind == "runtime" {
		writeCountFilesError = writeCountFiles(root, rows, nil)
	} else {
		writeCountFilesError = writeCountFiles(root, nil, rows)
	}
	if writeCountFilesError != nil {
		test.Fatal(writeCountFilesError)
	}
	actual, readError := os.ReadFile(filepath.Join(root, kind, filepath.FromSlash(key)+".counts"))
	if readError != nil {
		test.Fatal(readError)
	}
	expected, readError := os.ReadFile(countFile(kind, key))
	if readError != nil {
		test.Fatal(readError)
	}
	if !bytes.Equal(actual, expected) {
		test.Errorf(
			"regenerated counts differ in %s:\nrecorded: %sregenerated: %s",
			countFile(kind, key),
			expected,
			actual,
		)
	}
}

func TestPredicateCountsRegenerateDeterministically(test *testing.T) {
	test.Parallel()
	if *updateCounts {
		return
	}
	for key, row := range predicateCountsTable(test) {
		regenerateCountRecord(test, "predicates", key, row)
	}
}

func TestCountsRegenerateDeterministically(test *testing.T) {
	test.Parallel()
	if *updateCounts {
		return
	}
	// Every runtime row is also regenerated in its existing count-family shard.
	const key = "dedication/dedication.a"
	regenerateCountRecord(test, "runtime", key, countMeasured(test, key))
}
