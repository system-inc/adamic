package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/boundedrun"
)

func TestLiteralChildrenRequireRunBinding(t *testing.T) {
	t.Parallel()
	file := filepath.Join(t.TempDir(), "probe_test.go")
	for _, row := range []struct {
		source string
		want   string
	}{
		{`for _, row := range []struct{ source, name string }{{"not a child", "actual child"}} { t.Run(row.name,func(t *testing.T){for _, side := range []string{"emitted JavaScript"} { t.Log(side) }}) }`, "actual_child"},
		{`for _, side := range []string{"emitted JavaScript"} { t.Log(side) }`, ""},
		{`for _, name := range []string{"emitted JavaScript"} { if false { t.Run(name,func(t *testing.T){}) } }`, ""},
		{`for _, name := range []string{"emitted JavaScript"} { name="different"; t.Run(name,func(t *testing.T){}) }`, ""},
	} {
		if err := os.WriteFile(file, []byte("package probe\nimport \"testing\"\nfunc TestParent(t *testing.T){"+row.source+"}\n"), 0600); err != nil {
			t.Fatal(err)
		}
		names, err := literalChildren(file, "TestParent")
		if row.want == "" {
			if err == nil || len(names) > 0 {
				t.Fatal("phantom child was enumerated", names, err)
			}
		} else if err != nil || strings.Join(names, ",") != row.want {
			t.Fatal("wrong bound name", names, err)
		}
	}
}

// Not parallel: the planner uses the working directory and Git tree.
func TestPlannerDoesNotInventLintBackendChildren(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	t.Setenv("GOWORK", "off")
	for _, directory := range []string{"cmd/adamic-gate", "stage1/cohere/lint"} {
		if err := os.MkdirAll(directory, 0755); err != nil {
			t.Fatal(err)
		}
	}
	source := `package lint
import "testing"
func registeredNames() []string {return []string{"real registry mutant"}}
func TestMutants(t *testing.T) {
 for _, name := range registeredNames() {
  t.Run(name,func(t *testing.T){
   for _, side := range []struct{name string}{{"Node"},{"emitted JavaScript"},{"native"}} {t.Log(side.name)}
  })
 }
}
`
	for name, contents := range map[string]string{"go.mod": "module github.com/system-inc/adamic\n\ngo 1.27\n", timingPath: "{}\n", "stage1/cohere/lint/lint_test.go": source} {
		if err := os.WriteFile(name, []byte(contents), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if names, err := literalChildren("stage1/cohere/lint/lint_test.go", "TestMutants"); err == nil || len(names) > 0 {
		t.Fatal("backend loop confused with registration", names, err)
	}
	gitFixture(t, root, "init", "-q")
	gitFixture(t, root, "config", "user.name", "Gate fixture")
	gitFixture(t, root, "config", "user.email", "gate@example.invalid")
	gitFixture(t, root, "add", ".")
	gitFixture(t, root, "commit", "-qm", "Lint planner fixture")
	p, err := makePlan(15)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Units) != 1 || p.Units[0].Test != "TestMutants" {
		t.Fatal("planner invented a child", p.Units)
	}
	log := filepath.Join(t.TempDir(), "test.jsonl")
	cmd, release := boundedrun.Command(boundedrun.Build, "go", "test", "-count=1", "-json", "./stage1/cohere/lint")
	cmd.Env = append(os.Environ(), "ADAMIC_GATE_UNCACHED=1")
	err = runJSONCommand(cmd, log)
	release()
	if err != nil {
		t.Fatal(err)
	}
	results, _, _, err := readLog(log)
	if err != nil {
		t.Fatal(err)
	}
	events := map[string]bool{}
	for _, result := range results {
		events[result.Test] = true
	}
	for _, unit := range p.Units {
		if !events[unit.Test] {
			t.Fatal("planned child did not run", unit.Test)
		}
	}
	if !events["TestMutants/real_registry_mutant"] || events["TestMutants/emitted_JavaScript"] {
		t.Fatal("fixture did not reproduce registration", events)
	}
}
