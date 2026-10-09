package lint

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/stage1/cohere/lint/registry"
)

// Validate every descriptor before even a filtered test. No rule can hide behind a filter.
func TestMain(m *testing.M) {
	if _, err := registry.Generate("."); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	directory, err := os.MkdirTemp("", "lint-shared-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	sharedDirectory = directory
	code := m.Run()
	if err := os.RemoveAll(directory); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	os.Exit(code)
}

func TestRegistrationMutant(t *testing.T) {
	// This remains a valid descriptor and compiled rule: only its subscription is wrong.
	directory := mutant(t, `"DebuggerStatement"`, `"EmptyStatement"`, "rules/no-debugger/rule.json")
	source := ownedWitnesses(t, ".", "no-debugger")[0]
	path := manifest(t, []string{source + "\tno-debugger"})
	want := execute(t, "", goOracle(t), "--manifest", path).output
	for _, side := range []struct {
		name string
		run  execution
	}{
		{"Node", node(t, directory, path, false)},
		{"emitted JavaScript", emittedNode(t, directory, path, false)},
	} {
		if bytes.Equal(side.run.output, want) {
			t.Fatalf("listener omission survived on %s", side.name)
		}
		t.Logf("wrong descriptor kind caught on %s: %s", side.name, difference(side.run.output, want))
	}
}

func TestFactoryHooks(t *testing.T) {
	factoryHooksProducts(t)
}

// Raw corpus text stays outside the project's TypeScript module graph.
func ownedWitnesses(t *testing.T, directory, slug string) []string {
	t.Helper()
	paths, err := registry.Witnesses(filepath.Join(directory, "rules", slug))
	if err != nil {
		t.Fatal(err)
	}
	var sources []string
	for index, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		target := filepath.Join(t.TempDir(), fmt.Sprintf("%s-%d%s", slug, index, filepath.Ext(strings.TrimSuffix(path, ".txt"))))
		// A witness in a directory under testdata is linted at that relative path, for a rule that judges it.
		relative, err := filepath.Rel(filepath.Join(directory, "rules", slug, "testdata"), path)
		if err != nil {
			t.Fatal(err)
		}
		if filepath.Dir(relative) != "." {
			target = filepath.Join(t.TempDir(), strings.TrimSuffix(relative, ".txt"))
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				t.Fatal(err)
			}
		}
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

// ownedWitnessRows returns a manifest row selecting the rule for each of its witnesses. A witness may carry
// the rule's options beside it, as foo.options.json next to foo.ts.txt: a rule that reports nothing by
// default, such as one that bans only the types it is configured with, can only witness a finding with
// options. They go in the row's options field, as a captured upstream case's do, so the oracle hands them to
// the rule's adapter and the port reads the same settings. The "all" row stays unconfigured, because there
// the options would be every rule's.
func ownedWitnessRows(t *testing.T, directory string, d registry.Descriptor) []string {
	t.Helper()
	paths, err := registry.Witnesses(filepath.Join(directory, "rules", d.Slug))
	if err != nil {
		t.Fatal(err)
	}
	sources := ownedWitnesses(t, directory, d.Slug)
	var rows []string
	for index, path := range paths {
		row := sources[index] + "\t" + d.Name
		sidecar := strings.TrimSuffix(strings.TrimSuffix(path, ".txt"), filepath.Ext(strings.TrimSuffix(path, ".txt"))) +
			".options.json"
		data, err := os.ReadFile(sidecar)
		if err == nil {
			var options any
			if err := json.Unmarshal(data, &options); err != nil {
				t.Fatalf("%s: %v", sidecar, err)
			}
			compact, err := json.Marshal(options)
			if err != nil {
				t.Fatal(err)
			}
			row += "\t\t\tfalse\t" + string(compact)
		} else if !os.IsNotExist(err) {
			t.Fatal(err)
		}
		rows = append(rows, row)
	}
	return rows
}

func TestNestedOutsideModuleCopy(t *testing.T) {
	directory := mutant(t, "", "")
	original := `import { written } from '../../../../typescript/parser/nodes.ts';
console.log(written('copied'));
`
	source := rewritePortImports(t, "rules/probe/rule.ts", original)
	if err := os.WriteFile(filepath.Join(directory, "main.ts"), []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	path := manifest(t, nil)
	for _, side := range []struct {
		name string
		run  execution
	}{
		{"Node", node(t, directory, path, false)},
		{"native", execute(t, "", buildPort(t, directory, true))},
	} {
		if string(side.run.output) != "copied\n" {
			t.Fatalf("outside module copy on %s: %q", side.name, side.run.output)
		}
	}
	// Put the old root-only rewrite back: a nested rule keeps two unwanted parents.
	parserDirectory, err := filepath.Abs("../../typescript")
	if err != nil {
		t.Fatal(err)
	}
	bad := strings.ReplaceAll(original, "../../typescript", parserDirectory)
	main := filepath.Join(directory, "main.ts")
	if err := os.WriteFile(main, []byte(bad), 0644); err != nil {
		t.Fatal(err)
	}
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	output, err := os.CreateTemp(t.TempDir(), "bad-copy-")
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command("node", "--disable-warning=ExperimentalWarning", runner, main)
	command.Stdout = output
	command.Stderr = output
	runError := command.Run()
	if err := output.Close(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(output.Name())
	if err != nil {
		t.Fatal(err)
	}
	if runError == nil || !bytes.Contains(data, []byte("ERR_MODULE_NOT_FOUND")) {
		t.Fatalf("root-only import mutant survived on Node: %v %s", runError, data)
	}
	if _, err := load.Load([]string{main}); err == nil {
		t.Fatal("root-only import mutant survived native module loading")
	}
	t.Log("root-only import rewrite mutant caught by Node and native module loading")
}

func TestDecodedOptionsAndMutant(t *testing.T) {
	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	fixture := filepath.Join(t.TempDir(), "catch.ts")
	if err := os.WriteFile(fixture, []byte("try { work(); } catch(e) {}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	path := manifest(t, []string{fixture + "\tno-empty\t\t\tfalse\t{\"AllowEmptyCatch\":true}"})
	oracle := goOracle(t)
	compare(t, oracle, buildPort(t, directory, true), directory, path)
	count := execute(t, "", oracle, "--manifest", path, "--count")
	if string(count.output) != "0\n" {
		t.Fatal("JSON catch option did not override the legacy default")
	}
	changed := mutant(t, "'allowemptycatch'", "'ignored-allowemptycatch'", "main.ts")
	want := execute(t, "", oracle, "--manifest", path).output
	for _, side := range []struct {
		name string
		run  execution
	}{
		{"Node", node(t, changed, path, false)},
		{"emitted JavaScript", emittedNode(t, changed, path, false)},
	} {
		if bytes.Equal(side.run.output, want) {
			t.Fatalf("ignored decoded-option mutant survived on %s", side.name)
		}
		t.Logf("ignored decoded-option mutant caught on %s: %s", side.name, difference(side.run.output, want))
	}
}
