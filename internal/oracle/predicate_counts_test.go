package oracle

import (
	"fmt"
	"path/filepath"
	"sort"
	"testing"
)

func predicateCountKeys() ([]string, error) {
	var keys []string
	for _, extension := range []string{"a", "ts"} {
		paths, err := filepath.Glob(filepath.Join(repository, "internal/lower/testdata/predicates/overload_*."+extension))
		if err != nil {
			return nil, err
		}
		for _, path := range paths {
			relative, err := filepath.Rel(repository, path)
			if err != nil {
				return nil, err
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

func predicateCountsTable(t *testing.T) map[string]string {
	t.Helper()
	keys, err := predicateCountKeys()
	if err != nil {
		t.Fatal(err)
	}
	table := map[string]string{}
	for _, relative := range keys {
		path := filepath.Join(repository, relative)

		program, err := lowered(t, path)
		if err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		counts := program.PredicateChecks
		proven, checked, unobservable := 0, 0, 0
		for _, site := range counts.Sites {
			if site.Where == "" || site.Function == "" || site.Overload < 1 {
				t.Fatalf("incomplete predicate site: %+v", site)
			}
			seen := map[string]bool{}
			for _, direction := range site.Directions {
				if (direction.Direction != "true" && direction.Direction != "false") || seen[direction.Direction] || direction.Reason == "" {
					t.Fatalf("invalid direction: %+v", direction)
				}
				seen[direction.Direction] = true
				switch direction.Status {
				case "proven":
					proven++
				case "checked":
					checked++
				case "unobservable":
					proven++
					unobservable++
				default:
					t.Fatalf("unknown predicate status: %+v", direction)
				}
			}
		}
		if counts.Proven != proven || counts.Checked != checked || counts.Unobservable != unobservable {
			t.Fatalf("%s: direction records (%d,%d,%d) disagree with aggregate %+v", path, proven, checked, unobservable, counts)
		}
		table[relative] = fmt.Sprintf("| %s | %d | %d | %d | %d |", relative, len(counts.Sites), proven, checked, unobservable)
	}
	return table
}

func TestPredicateDirectionCountsAreRecorded(t *testing.T) {
	t.Parallel()
	// TestMain owns committed writes during -update-counts.
	if *updateCounts {
		return
	}
	for key, got := range predicateCountsTable(t) {
		want, err := readCountRow(countsDirectory, "predicates", key)
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Errorf("predicate counts changed in %s; update with -update-counts:\nrecorded: %s\nmeasured: %s", countFile("predicates", key), want, got)
		}
	}
}
