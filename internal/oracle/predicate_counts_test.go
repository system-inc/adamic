package oracle

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const predicateCountsHeader = `
## Predicate direction counts

Each emitted overload predicate call reports its true and false directions (assertions
have only a true direction). Unobservable directions have no narrowed read in their
flow region; they are included in Proven and also reported separately. These are
compiler proof counts, independent of the runtime allocation counts above.

| Fixture | Call sites | Proven | Checked | Unobservable |
|---|---:|---:|---:|---:|
`

func predicateCountsTable(t *testing.T) string {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(repository, "internal/lower/testdata/predicates/overload_*.a"))
	if err != nil || len(paths) == 0 {
		t.Fatalf("predicate fixtures: %v", err)
	}
	var table strings.Builder
	table.WriteString(predicateCountsHeader)
	for _, path := range paths {
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
		relative, err := filepath.Rel(repository, path)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Fprintf(&table, "| %s | %d | %d | %d | %d |\n", filepath.ToSlash(relative), len(counts.Sites), proven, checked, unobservable)
	}
	return table.String()
}

func TestPredicateDirectionCountsAreRecorded(t *testing.T) {
	t.Parallel()
	// The full count writer owns counts.md during -update-counts.
	if *updateCounts {
		return
	}
	recorded, err := os.ReadFile(countsPath)
	if err != nil {
		t.Fatal(err)
	}
	index := strings.Index(string(recorded), predicateCountsHeader)
	if index < 0 {
		t.Fatal("missing predicate direction table in counts.md")
	}
	if got := predicateCountsTable(t); string(recorded[index:]) != got {
		t.Fatalf("predicate counts changed; update counts.md with the full count writer:\n%s", got)
	}
}
