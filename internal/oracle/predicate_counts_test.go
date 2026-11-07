package oracle

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func predicateCountKeys() ([]string, error) {
	var keys []string
	for _, extension := range []string{"a", "ts"} {
		// Reads: internal/lower/testdata/*
		paths, globError := filepath.Glob(
			filepath.Join(repository, "internal/lower/testdata/predicates/overload_*."+extension),
		)
		if globError != nil {
			return nil, globError
		}
		for _, path := range paths {
			relative, relativePathError := filepath.Rel(repository, path)
			if relativePathError != nil {
				return nil, relativePathError
			}
			keys = append(keys, filepath.ToSlash(relative))
		}
	}
	sort.Strings(keys)
	if len(keys) == 0 {
		return nil, fmt.Errorf("no predicate fixtures")
	}
	return keys, nil
}

func predicateCountsTable(test *testing.T) map[string]string {
	test.Helper()
	keys, predicateCountKeysError := predicateCountKeys()
	if predicateCountKeysError != nil {
		test.Fatal(predicateCountKeysError)
	}
	table := map[string]string{}
	for _, relative := range keys {
		path := filepath.Join(repository, relative)

		source, readError := os.ReadFile(path)
		if readError != nil {
			test.Fatal(readError)
		}
		checkedPath := filepath.Join(test.TempDir(), "main.ts")
		if writeError := os.WriteFile(checkedPath, source, 0644); writeError != nil {
			test.Fatal(writeError)
		}
		program, loweredError := lowered(test, checkedPath)
		if loweredError != nil {
			test.Fatalf("%s: %v", path, loweredError)
		}
		counts := program.PredicateChecks
		proven, checked, unobservable := 0, 0, 0
		for _, site := range counts.Sites {
			if site.Where == "" || site.Function == "" || site.Overload < 1 {
				test.Fatalf("incomplete predicate site: %+v", site)
			}
			seen := map[string]bool{}
			for _, direction := range site.Directions {
				if (direction.Direction != "true" &&
					direction.Direction != "false" &&
					direction.Direction != "asserts") || seen[direction.Direction] || direction.Reason == "" {
					test.Fatalf("invalid direction: %+v", direction)
				}
				seen[direction.Direction] = true
				switch direction.Status {
				case "proven":
					proven++
				case "checked":
					checked++
				case "unobservable":
					unobservable++
				default:
					test.Fatalf("unknown predicate status: %+v", direction)
				}
			}
		}
		if counts.Proven != proven || counts.Checked != checked || counts.Unobservable != unobservable {
			test.Fatalf(
				"%s: direction records (%d,%d,%d) disagree with aggregate %+v",
				path,
				proven,
				checked,
				unobservable,
				counts,
			)
		}
		table[relative] = fmt.Sprintf(
			"| %s | %d | %d | %d | %d |",
			relative,
			len(counts.Sites),
			proven,
			checked,
			unobservable,
		)
	}
	return table
}

func TestPredicateDirectionCountsAreRecorded(test *testing.T) {
	test.Parallel()
	// TestMain owns committed writes during -update-counts.
	if *updateCounts {
		return
	}
	for key, got := range predicateCountsTable(test) {
		want, readCountRowError := readCountRow(countsDirectory, "predicates", key)
		if readCountRowError != nil {
			test.Fatal(readCountRowError)
		}
		if got != want {
			test.Errorf(
				"predicate counts changed in %s; update with -update-counts:\nrecorded: %s\nmeasured: %s",
				countFile("predicates", key),
				want,
				got,
			)
		}
	}
}
