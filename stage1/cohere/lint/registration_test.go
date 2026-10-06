package lint

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/stage1/cohere/lint/registry"
)

// Validate every descriptor before even a filtered test. No rule can hide behind a filter.
func TestMain(m *testing.M) {
	if _, err := registry.Generate("."); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}

func TestOwnedWitnesses(t *testing.T) {
	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	oracle := goOracle(t)
	var rows []string
	for _, d := range prepareRegistry(t, ".") {
		paths := ownedWitnesses(t, directory, d.Slug)
		for _, path := range paths {
			selected := path + "\t" + d.Name
			answer := execute(t, "", oracle, "--manifest", manifest(t, []string{selected}), "--count")
			if string(answer.output) == "0\n" {
				t.Fatalf("%s witness reports no findings", d.Name)
			}
			rows = append(rows, selected, path+"\tall")
		}
	}
	path := manifest(t, rows)
	compare(t, oracle, buildPort(t, directory, true), directory, path)
}

func TestRegistrationMutant(t *testing.T) {
	// This remains a valid descriptor and compiled rule: only its subscription is wrong.
	directory := mutant(t, `"DebuggerStatement"`, `"EmptyStatement"`)
	source := ownedWitnesses(t, ".", "no-debugger")[0]
	path := manifest(t, []string{source + "\tno-debugger"})
	want := execute(t, "", goOracle(t), "--manifest", path).output
	for _, side := range []struct {
		name string
		run  execution
	}{
		{"Node", node(t, directory, path, false)},
		{"native", execute(t, "", buildPort(t, directory, true), "--manifest", path)},
	} {
		if bytes.Equal(side.run.output, want) {
			t.Fatalf("listener omission survived on %s", side.name)
		}
		t.Logf("wrong descriptor kind caught on %s: %s", side.name, difference(side.run.output, want))
	}
}

func TestFactoryHooks(t *testing.T) {
	directory := mutant(t, "", "")
	module := filepath.Join(directory, "rules/no-debugger/rule.ts")
	data, err := os.ReadFile(module)
	if err != nil {
		t.Fatal(err)
	}
	source := strings.Replace(string(data), "    visit(index: number, parent: number): void {", `    ready = false;
    prepare(root: number): void { this.ready = true; console.log('prepare'); }
    finish(root: number): void { console.log('finish'); }
    visit(index: number, parent: number): void {
        if(!this.ready) { panic('visit before prepare'); }
        console.log('visit');`, 1)
	source = "import { panic } from 'adamic';\n" + source
	if err := os.WriteFile(module, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	descriptor := filepath.Join(directory, "rules/no-debugger/rule.json")
	data, err = os.ReadFile(descriptor)
	if err != nil {
		t.Fatal(err)
	}
	var options map[string]any
	if err := json.Unmarshal(data, &options); err != nil {
		t.Fatal(err)
	}
	options["prepare"] = "prepare"
	options["finish"] = "finish"
	write := func() {
		data, err := json.Marshal(options)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(descriptor, data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	write()
	fixture := filepath.Join(t.TempDir(), "hooks.ts")
	if err := os.WriteFile(fixture, []byte("debugger;\n"), 0644); err != nil {
		t.Fatal(err)
	}
	path := manifest(t, []string{fixture + "\tno-debugger"})
	expected := []byte("case 0\nprepare\nvisit\nfinish\n")
	run := func(mutated bool) {
		for _, side := range []struct {
			name string
			run  execution
		}{
			{"Node", node(t, directory, path, false)},
			{"native", execute(t, "", buildPort(t, directory, true), "--manifest", path)},
		} {
			matches := bytes.HasPrefix(side.run.output, expected)
			if matches == mutated {
				t.Fatalf("hook sequence on %s (mutant=%t): %s", side.name, mutated, side.run.output)
			}
			if mutated {
				t.Logf("finish-hook omission caught on %s", side.name)
			}
		}
	}
	run(false)
	delete(options, "finish")
	write()
	run(true)
}

// Raw corpus text stays outside the project's TypeScript module graph.
func ownedWitnesses(t *testing.T, directory, slug string) []string {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(directory, "rules", slug, "testdata", "*.ts.txt"))
	if err != nil {
		t.Fatal(err)
	}
	var sources []string
	for index, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		target := filepath.Join(t.TempDir(), fmt.Sprintf("%s-%d.ts", slug, index))
		if err := os.WriteFile(target, data, 0644); err != nil {
			t.Fatal(err)
		}
		sources = append(sources, target)
	}
	if len(sources) == 0 {
		t.Fatalf("%s has no witnesses", slug)
	}
	return sources
}
