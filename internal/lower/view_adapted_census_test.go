package lower

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
)

// Measure production entry compilation without bypassing rejected checker input.
// Each cast is attributed to its owning file's actual import closure.
func TestLazyViewAdaptedCensus(t *testing.T) {
	root := os.Getenv("LAZY_ADAPTED_ROOT")
	if root == "" {
		t.Skip("set LAZY_ADAPTED_ROOT, LAZY_ADAPTED_SITES and LAZY_ADAPTED_OUTPUT")
	}
	type site struct {
		File, Kind   string
		Line, Column int
	}
	data, err := os.ReadFile(os.Getenv("LAZY_ADAPTED_SITES"))
	if err != nil {
		t.Fatal(err)
	}
	var sites []site
	if err := json.Unmarshal(data, &sites); err != nil {
		t.Fatal(err)
	}
	type entry struct {
		File        string
		Sites       map[string]int
		Stop        string
		Diagnostics []string
	}
	counts := map[string]int{"tagged": 0, "untagged": 0}
	compiled := map[string]int{"tagged": 0, "untagged": 0}
	byFile := map[string]*entry{}
	for _, s := range sites {
		if s.Kind != "tagged" && s.Kind != "untagged" {
			continue
		}
		counts[s.Kind]++
		if byFile[s.File] == nil {
			byFile[s.File] = &entry{File: s.File, Sites: map[string]int{}}
		}
		byFile[s.File].Sites[s.Kind]++
	}
	var roots []string
	err = filepath.WalkDir(filepath.Join(root, "src/compiler"), func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".ts") && !strings.HasSuffix(path, ".d.ts") {
			roots = append(roots, path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	diagnostics := func(err error) []string {
		if err == nil {
			return nil
		}
		var checked *load.CheckError
		if errors.As(err, &checked) {
			return checked.Diagnostics
		}
		return []string{err.Error()}
	}
	_, err = load.Load(roots)
	projectDiagnostics := diagnostics(err)
	var names []string
	for name := range byFile {
		names = append(names, name)
	}
	sort.Strings(names)
	var entries []*entry
	for _, name := range names {
		e := byFile[name]
		program, failure := load.Load([]string{filepath.Join(root, name)})
		e.Stop = "checker"
		if failure == nil {
			e.Stop = "lowering"
			_, failure = Lower(context.Background(), program)
			if failure == nil {
				t.Fatalf("entry %s reached IR: add both backend builds before counting it compiled", name)
			}
		}
		e.Diagnostics = diagnostics(failure)
		entries = append(entries, e)
		t.Logf("%s: %s, %d diagnostics, sites=%v", name, e.Stop, len(e.Diagnostics), e.Sites)
	}
	output := struct {
		Inventory, ProductionCompiled map[string]int
		ProjectDiagnostics            []string
		Entries                       []*entry
		Limits                        string
	}{counts, compiled, projectDiagnostics, entries, "Production compilation uses unchanged Adamic options and each entry's real import closure. Checker-rejected entries have no executable IR or shared allocation-flow graph; remaining read-family reachability is unmeasurable, not zero. No diagnostic bypass or partial lowering."}
	data, err = json.MarshalIndent(output, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(os.Getenv("LAZY_ADAPTED_OUTPUT"), append(data, '\n'), 0600); err != nil {
		t.Fatal(err)
	}
	t.Logf("inventory=%v compiled=%v project diagnostics=%d", counts, compiled, len(projectDiagnostics))
}
