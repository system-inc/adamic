package lint

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInheritedAllSelection(t *testing.T) {
	d := selectedDescriptor(t, "no-var")
	rows, owned := ruleRows(t, d)
	inherited := stableRows(t, inheritedRuleRows(generated(t), d.Name))
	available := map[string]bool{}
	for _, row := range rows {
		available[row] = true
	}
	for _, row := range inherited {
		if !available[row] {
			t.Fatalf("original inherited manifest row lost: %s", row)
		}
	}
	mutated := selectedPortMutation(t, d, "const node = this.context.node(index);", "if(this.context.selected === 'all') { return; }\n        const node = this.context.node(index);", filepath.Join("rules", d.Slug, d.Module))
	oracle := goOracleFrom(t, mutated)
	selected := manifest(t, owned)
	if diff := difference(node(t, mutated, selected, false).output, execute(t, "", oracle, "--manifest", selected).output); diff != "" {
		t.Fatalf("all-selector mutant caught outside inherited all rows: %s", diff)
	}
	path := manifest(t, inherited)
	if bytes.Equal(node(t, mutated, path, false).output, execute(t, "", oracle, "--manifest", path).output) {
		t.Fatal("inherited all-selector mutant survived")
	}
	t.Log("inherited all-selector mutant caught; selected witnesses remain identical")
}

// Registration scope must not remove a module imported by the selected rule.
func TestSelectedRegistrationKeepsImportedHelpers(t *testing.T) {
	d := selectedDescriptor(t, "no-var")
	directory := selectedPort(t, d)
	source := filepath.Join(directory, "rules", d.Slug)
	extra := filepath.Join(directory, "rules", "unused-probe")
	err := filepath.WalkDir(source, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		destination := filepath.Join(extra, relative)
		if entry.IsDir() {
			return os.MkdirAll(destination, 0755)
		}
		return os.WriteFile(destination, []byte(lintBytes(t, path)), 0644)
	})
	if err != nil {
		t.Fatal(err)
	}
	descriptor := filepath.Join(extra, "rule.json")
	var fields map[string]any
	if err := json.Unmarshal([]byte(lintBytes(t, descriptor)), &fields); err != nil {
		t.Fatal(err)
	}
	fields["name"] = "unused-probe"
	fields["oracle"] = "oracleUnusedProbe"
	delete(fields, "order")
	encoded, err := json.Marshal(fields)
	if err != nil {
		t.Fatal(err)
	}
	lintPublish(t, descriptor, encoded, 0644)
	adapter := filepath.Join(extra, "oracle.go")
	lintPublish(t, adapter, []byte(strings.ReplaceAll(lintBytes(t, adapter), "oracleNoVar", "oracleUnusedProbe")), 0644)
	helper := filepath.Join(extra, d.Module)
	lintPublish(t, helper, []byte(lintBytes(t, helper)+"\nexport const cacheHelper = 'first';\n"), 0644)
	target := filepath.Join(source, d.Module)
	lintPublish(t, target, []byte(lintBytes(t, target)+"\nimport { cacheHelper } from '../unused-probe/"+d.Module+"';\nconsole.log(cacheHelper);\n"), 0644)
	prepareRegistry(t, directory)
	if strings.Contains(lintBytes(t, filepath.Join(directory, ".generated/registry.ts")), "unused-probe") {
		t.Fatal("unselected registration entered the generated registry")
	}
	before := lintModules(t, filepath.Join(directory, "main.ts"))
	lintChange(t, helper, "cacheHelper = 'first'", "cacheHelper = 'other'")
	if before == lintModules(t, filepath.Join(directory, "main.ts")) {
		t.Fatal("imported unselected helper bytes did not invalidate the graph")
	}
	t.Log("unselected listener omitted; its imported helper retained and keyed")
}

func TestSelectedRuleUncachedAnswers(t *testing.T) {
	// Not parallel: the bypass changes process-wide execution state.
	t.Setenv("ADAMIC_GATE_UNCACHED", "0")
	d := selectedDescriptor(t, "no-var")
	directory := selectedPort(t, d)
	rows, _ := ruleRows(t, d)
	rows = append(rows, stableRows(t, lintCapture(t, ".", d))...)
	path := manifest(t, rows)
	want := compare(t, goOracleFrom(t, directory), buildPort(t, directory, true), directory, path)
	t.Setenv("ADAMIC_GATE_UNCACHED", "1")
	got := compare(t, goOracleFrom(t, directory), buildPort(t, directory, true), directory, path)
	if !bytes.Equal(got, want) {
		t.Fatal("uncached selected check changed answer bytes")
	}
	t.Logf("cached and ADAMIC_GATE_UNCACHED=1 answers byte-identical: %d bytes", len(got))
}
