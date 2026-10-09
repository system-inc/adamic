package oracle

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

const predicateCountsHeader = `
## Predicate direction counts

Each overload witness is loaded from a temporary .ts copy: checked admission is
explicit, while its stored .a source remains a Node witness. True directions are
checked once; false directions are checked where the checker excludes part of the
incoming type. Assertions have an asserts direction. Unobservable directions are
reported separately and are not body proofs. These are compiler proof counts,
independent of the runtime allocation counts above.

| Fixture | Call sites | Proven | Checked | Unobservable |
|---|---:|---:|---:|---:|
`

func predicateCountsTable(t *testing.T) string {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(repository, "internal/lower/testdata/predicates/overload_*.a"))
	if err != nil || len(paths) == 0 {
		t.Fatalf("predicate fixtures: %v", err)
	}
	typeScriptPaths, err := filepath.Glob(filepath.Join(repository, "internal/lower/testdata/predicates/overload_*.ts"))
	if err != nil {
		t.Fatal(err)
	}
	paths = append(paths, typeScriptPaths...)
	sort.Strings(paths)
	var table strings.Builder
	table.WriteString(predicateCountsHeader)
	for _, path := range paths {

		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		checkedPath := filepath.Join(t.TempDir(), "main.ts")
		if err = os.WriteFile(checkedPath, source, 0644); err != nil {
			t.Fatal(err)
		}
		program, err := lowered(t, checkedPath)
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
				if (direction.Direction != "true" && direction.Direction != "false" && direction.Direction != "asserts") || seen[direction.Direction] || direction.Reason == "" {
					t.Fatalf("invalid direction: %+v", direction)
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
	for _, fixture := range conditionAssertionFixtures {
		program, err := lowered(t, conditionAssertionInput(t, fixture.name, true))
		if err != nil {
			t.Fatal(err)
		}
		counts := program.PredicateChecks
		if len(counts.Sites) != 1 || len(counts.Sites[0].Directions) != 1 {
			t.Fatalf("incomplete condition site: %+v", counts)
		}
		direction := counts.Sites[0].Directions[0]
		status := "checked"
		if fixture.proven {
			status = "proven"
		}
		if direction.Direction != "asserts" || direction.Status != status || direction.Reason == "" || counts.Proven+counts.Checked != 1 || counts.Unobservable != 0 {
			t.Fatalf("invalid condition direction: %+v", counts)
		}
		fmt.Fprintf(&table, "| internal/lower/testdata/condition_assertions/%s.a (.ts checked mode) | %d | %d | %d | %d |\n", fixture.name, len(counts.Sites), counts.Proven, counts.Checked, counts.Unobservable)
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
