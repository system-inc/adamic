package lint

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/toolchain"
	"github.com/system-inc/adamic/stage1/cohere/lint/registry"
)

// Validate every descriptor before even a filtered test. No rule can hide behind a filter.
// Not parallel: TestMain owns the shared temporary directory and registry initialization.
func TestMain(m *testing.M) {
	// The committed registry is checked, never rewritten: the tests don't write the tree (#k01pgng).
	if _, err := registry.Stale("."); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	directory, err := os.MkdirTemp("", "lint-shared-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	sharedDirectory = directory
	if err := decodedOptionsBeforeTests(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.RemoveAll(directory)
		os.Exit(1)
	}
	code := m.Run()
	if err := os.RemoveAll(directory); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	os.Exit(code)
}

func TestRegistrationMutant(test *testing.T) {
	test.Parallel()
	// This remains a valid descriptor and compiled rule: only its subscription is wrong.
	directory := mutant(test, `"DebuggerStatement"`, `"EmptyStatement"`, "rules/no-debugger/rule.json")
	source := ownedWitnesses(test, ".", "no-debugger")[0]
	path := manifest(test, []string{source + "\tno-debugger"})
	want := execute(test, "", lintGoOracleProduct(test), "--manifest", path).output
	for _, side := range []struct {
		name string
		run  execution
	}{
		{"Node", node(test, directory, path, false)},
		{"emitted JavaScript", emittedNode(test, directory, path, false)},
	} {
		if bytes.Equal(side.run.output, want) {
			test.Fatalf("listener omission survived on %s", side.name)
		}
		test.Logf("wrong descriptor kind caught on %s: %s", side.name, difference(side.run.output, want))
	}
}

// Not parallel: initializes the shared immutable hook fixtures before the parallel shards.
func TestFactoryHooks_Setup(test *testing.T) {
	factoryHooksSetup(test)
}

// Raw corpus text stays outside the project's TypeScript module graph.
func ownedWitnesses(test *testing.T, directory, slug string) []string {
	test.Helper()
	paths, err := registry.Witnesses(filepath.Join(directory, "rules", slug))
	if err != nil {
		test.Fatal(err)
	}
	var sources []string
	for index, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			test.Fatal(err)
		}
		target := filepath.Join(
			test.TempDir(),
			fmt.Sprintf("%s-%d%s", slug, index, filepath.Ext(strings.TrimSuffix(path, ".txt"))),
		)
		// A witness in a directory under testdata is linted at that relative path, for a rule that judges it.
		relative, err := filepath.Rel(filepath.Join(directory, "rules", slug, "testdata"), path)
		if err != nil {
			test.Fatal(err)
		}
		if filepath.Dir(relative) != "." {
			target = filepath.Join(test.TempDir(), strings.TrimSuffix(relative, ".txt"))
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				test.Fatal(err)
			}
		}
		if err := os.WriteFile(target, data, 0644); err != nil {
			test.Fatal(err)
		}
		sources = append(sources, target)
	}
	if len(sources) == 0 {
		test.Fatalf("%s has no witnesses", slug)
	}
	return sources
}

// ownedWitnessRows returns a manifest row selecting the rule for each of its witnesses. A witness may carry
// the rule's options beside it, as foo.options.json next to foo.ts.txt: a rule that reports nothing by
// default, such as one that bans only the types it is configured with, can only witness a finding with
// options. They go in the row's options field, as a captured upstream case's do, so the oracle hands them to
// the rule's adapter and the port reads the same settings. The "all" row stays unconfigured, because there
// the options would be every rule's.
func ownedWitnessRows(test *testing.T, directory string, d registry.Descriptor) []string {
	test.Helper()
	paths, err := registry.Witnesses(filepath.Join(directory, "rules", d.Slug))
	if err != nil {
		test.Fatal(err)
	}
	sources := ownedWitnesses(test, directory, d.Slug)
	var rows []string
	for index, path := range paths {
		row := sources[index] + "\t" + d.Name
		sidecar := strings.TrimSuffix(
			strings.TrimSuffix(path, ".txt"),
			filepath.Ext(strings.TrimSuffix(path, ".txt")),
		) +
			".options.json"
		data, err := os.ReadFile(sidecar)
		if err == nil {
			var options any
			if err := json.Unmarshal(data, &options); err != nil {
				test.Fatalf("%s: %v", sidecar, err)
			}
			compact, err := json.Marshal(options)
			if err != nil {
				test.Fatal(err)
			}
			row += "\t\t\tfalse\t" + string(compact)
		} else if !os.IsNotExist(err) {
			test.Fatal(err)
		}
		rows = append(rows, row)
	}
	return rows
}

func TestNestedOutsideModuleCopy(test *testing.T) {
	test.Parallel()
	directory := mutant(test, "", "")
	original := `import { written } from '../../../../typescript/parser/nodes.ts';
console.log(written('copied'));
`
	source := rewritePortImports(test, "rules/probe/rule.ts", original)
	if err := os.WriteFile(filepath.Join(directory, "main.ts"), []byte(source), 0644); err != nil {
		test.Fatal(err)
	}
	path := manifest(test, nil)
	for _, side := range []struct {
		name string
		run  execution
	}{
		{"Node", node(test, directory, path, false)},
		{"native", execute(test, "", buildPort(test, directory, true))},
	} {
		if string(side.run.output) != "copied\n" {
			test.Fatalf("outside module copy on %s: %q", side.name, side.run.output)
		}
	}
	// Put the old root-only rewrite back: a nested rule keeps two unwanted parents.
	parserDirectory, err := filepath.Abs("../../typescript")
	if err != nil {
		test.Fatal(err)
	}
	bad := strings.ReplaceAll(original, "../../typescript", parserDirectory)
	main := filepath.Join(directory, "main.ts")
	if err := os.WriteFile(main, []byte(bad), 0644); err != nil {
		test.Fatal(err)
	}
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		test.Fatal(err)
	}
	output, err := os.CreateTemp(test.TempDir(), "bad-copy-")
	if err != nil {
		test.Fatal(err)
	}
	command := toolchain.Command("node", "--disable-warning=ExperimentalWarning", runner, main)
	command.Stdout = output
	command.Stderr = output
	runError := command.Run()
	if err := output.Close(); err != nil {
		test.Fatal(err)
	}
	data, err := os.ReadFile(output.Name())
	if err != nil {
		test.Fatal(err)
	}
	if runError == nil || !bytes.Contains(data, []byte("ERR_MODULE_NOT_FOUND")) {
		test.Fatalf("root-only import mutant survived on Node: %v %s", runError, data)
	}
	if _, err := load.Load([]string{main}); err == nil {
		test.Fatal("root-only import mutant survived native module loading")
	}
	test.Log("root-only import rewrite mutant caught by Node and native module loading")
}
